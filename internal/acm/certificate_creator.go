package acm

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	r53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
	myroute53 "ziniki.org/deployer/modules/aws/internal/route53"
)

type certificateCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	teardown corebottom.TearDown
	props    map[driverbottom.Identifier]driverbottom.Expr

	client  *acm.Client
	route53 *route53.Client
}

func (acmc *certificateCreator) Loc() *errorsink.Location {
	return acmc.loc
}

func (acmc *certificateCreator) ShortDescription() string {
	return "aws.CertificateManager.Certificate[" + acmc.name + "]"
}

func (acmc *certificateCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.CertificateManager.Certificate[")
	iw.AttrsWhere(acmc)
	iw.TextAttr("named", acmc.name)
	iw.EndAttrs()
}

func (acmc *certificateCreator) DetermineInitialState(pres driverbottom.ValuePresenter) {
	ae := acmc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := ae.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	acmc.client = awsEnv.ACMClient()
	acmc.route53 = awsEnv.Route53Client()

	certs := acmc.findCertificatesFor(acmc.name)
	if len(certs) == 0 {
		log.Printf("there were no certs found for %s\n", acmc.name)
		pres.NotFound()
	} else {
		model := NewCertificateModel(acmc.loc)
		model.name = acmc.name

		// log.Printf("found %d certs for %s\n", len(certs), acmc.name)
		model.arn = certs[0]

		// acmc.describeCertificate(acmc.arn)
		pres.Present(model)
	}
}

func (acmc *certificateCreator) DetermineDesiredState(pres driverbottom.ValuePresenter) {
	model := NewCertificateModel(acmc.loc)
	for k, p := range acmc.props {
		v := acmc.tools.Storage.Eval(p)
		switch k.Id() {
		case "Domain":
			domain, ok := v.(myroute53.ExportedDomain)
			if !ok {
				log.Fatalf("Domain did not point to a domain instance")
			}
			model.hzid = domain.HostedZoneId()
		case "SubjectAlternativeNames":
			san, ok := utils.AsStringList(v)
			if !ok {
				justString, ok := v.(string)
				if !ok {
					log.Fatalf("SubjectAlternativeNames must be a list of strings")
					return
				} else {
					san = []string{justString}
				}
			}
			model.sans = san
		case "ValidationMethod":
			meth, ok := utils.AsStringer(v)
			if !ok {
				log.Fatalf("ValidationMethod must be a string")
				return
			}
			model.validationMethod = meth
		default:
			log.Fatalf("certificate coin does not support a parameter %s\n", k.Id())
		}
	}
	pres.Present(model)
}

func (acmc *certificateCreator) UpdateReality() {
	found := acmc.tools.Storage.GetCoin(acmc.coin, corebottom.DETERMINE_INITIAL_MODE).(*certificateModel)

	if found != nil {
		log.Printf("certificate %s already existed for %s\n", found.arn, found.name)
		return
	}

	desired := acmc.tools.Storage.GetCoin(acmc.coin, corebottom.DETERMINE_DESIRED_MODE).(*certificateModel)

	created := NewCertificateModel(desired.loc)
	created.name = desired.name
	created.hzid = desired.hzid

	vm := types.ValidationMethod(desired.validationMethod.String())
	if vm == "" {
		vm = types.ValidationMethodDns
	}

	input := acm.RequestCertificateInput{DomainName: &acmc.name, ValidationMethod: vm}
	if len(desired.sans) > 0 {
		input.SubjectAlternativeNames = desired.sans
	}
	req, err := acmc.client.RequestCertificate(context.TODO(), &input)
	if err != nil {
		log.Printf("failed to request cert %s: %v\n", acmc.name, err)
	}
	log.Printf("requested cert for %s: %s\n", acmc.name, *req.CertificateArn)
	created.arn = *req.CertificateArn

	// Check if we still need to validate it ...

	// acmc.describeCertificate(*req.CertificateArn)

	utils.ExponentialBackoff(func() bool {
		return acmc.tryToValidateCert(created.arn, created.hzid)
	})

	acmc.tools.Storage.Bind(acmc.coin, created)
}

func (acmc *certificateCreator) TearDown() {
	found := acmc.tools.Storage.GetCoin(acmc.coin, corebottom.DETERMINE_INITIAL_MODE).(*certificateModel)

	if found == nil {
		log.Printf("no certificate existed for %s\n", acmc.name)
		return
	}

	log.Printf("you have asked to tear down certificate for %s (arn: %s) with mode %s\n", found.name, found.arn, acmc.teardown.Mode())
	switch acmc.teardown.Mode() {
	case "preserve":
		log.Printf("not deleting certificate %s because teardown mode is 'preserve'", found.name)
	case "delete":
		log.Printf("deleting certificate for %s with teardown mode 'delete'", found.name)
		DeleteCertificate(acmc.client, found.arn)
	default:
		log.Printf("cannot handle teardown mode '%s' for bucket %s", acmc.teardown.Mode(), found.name)
	}
}

func (acmc *certificateCreator) findCertificatesFor(name string) []string {
	ret := make([]string, 0)
	// TODO: need a loop on "NextToken"
	certs, err := acmc.client.ListCertificates(context.TODO(), &acm.ListCertificatesInput{})
	if err != nil {
		log.Fatalf("Failed to get certificates: %v\n", err)
	}
	for _, c := range certs.CertificateSummaryList {
		// log.Printf("have cert %s: %s\n", *c.DomainName, *c.CertificateArn)
		if *c.DomainName == name {
			log.Printf("found cert %s for %s\n", *c.CertificateArn, *c.DomainName)
			ret = append(ret, *c.CertificateArn)
		}
	}
	return ret
}

func (acmc *certificateCreator) DescribeCertificate(arn string) {
	cert, err := acmc.client.DescribeCertificate(context.TODO(), &acm.DescribeCertificateInput{CertificateArn: &arn})
	if err != nil {
		log.Fatalf("Failed to describe certificate %s: %v\n", arn, err)
	}
	fmt.Printf("cert arn: %s\n", arn)
	if cert.Certificate.DomainName != nil {
		fmt.Printf("cert domain: %s\n", *cert.Certificate.DomainName)
	}
	fmt.Printf("status: %s\n", cert.Certificate.Status)
	switch cert.Certificate.Status {
	case types.CertificateStatusFailed:
		fmt.Printf("failed: %s\n", cert.Certificate.FailureReason)
	case types.CertificateStatusIssued:
		fmt.Printf("until: %s\n", *cert.Certificate.NotAfter)
	case types.CertificateStatusPendingValidation:
		fmt.Printf("pending validation\n")
	}
}

func (acmc *certificateCreator) tryToValidateCert(arn string, hzid string) bool {
	cert, err := acmc.client.DescribeCertificate(context.TODO(), &acm.DescribeCertificateInput{CertificateArn: &arn})
	if err != nil {
		log.Fatalf("Failed to describe certificate %s: %v\n", arn, err)
	}

	switch cert.Certificate.Status {
	case types.CertificateStatusFailed:
		log.Printf("failed: %s\n", cert.Certificate.FailureReason)
		return true
	case types.CertificateStatusIssued:
		log.Printf("certificate issued until: %s\n", *cert.Certificate.NotAfter)
		return true
	case types.CertificateStatusPendingValidation:
		dns := make(map[string]string, 0)
		for _, x := range cert.Certificate.DomainValidationOptions {
			if x.ResourceRecord != nil && x.ResourceRecord.Name != nil && x.ResourceRecord.Value != nil {
				log.Printf("need %s => %s\n", *x.ResourceRecord.Name, *x.ResourceRecord.Value)
				dns[*x.ResourceRecord.Name] = *x.ResourceRecord.Value
			}
		}
		rrs, err := acmc.route53.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{HostedZoneId: &hzid})
		if err != nil {
			panic(err)
		}
		for _, r := range rrs.ResourceRecordSets {
			if r.Type == "CNAME" {
				log.Printf("already have %s %v\n", *r.Name, *r.ResourceRecords[0].Value)
				dns[*r.Name] = ""
			}
		}
		for k, v := range dns {
			if v != "" {
				log.Printf("creating %s to %s\n", k, v)
				var ttl int64 = 300
				changes := r53types.ResourceRecordSet{Name: &k, Type: "CNAME", TTL: &ttl, ResourceRecords: []r53types.ResourceRecord{{Value: &v}}}
				cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "CREATE", ResourceRecordSet: &changes}}}
				_, err := acmc.route53.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &hzid, ChangeBatch: &cb})
				if err != nil {
					panic(err)
				}
			}
		}

		return false
	default:
		panic("what is this? " + cert.Certificate.Status)
	}
}

func (acmc *certificateCreator) String() string {
	return fmt.Sprintf("EnsureBucket[%s:%s]", "" /* eacm.env.Region */, acmc.name)
}

func DeleteCertificate(client *acm.Client, arn string) {
	_, err := client.DeleteCertificate(context.TODO(), &acm.DeleteCertificateInput{CertificateArn: &arn})
	if err != nil {
		panic(err)
	}
	log.Printf("I think the certificate was deleted because no error was reported")
}

var _ corebottom.Ensurable = &certificateCreator{}

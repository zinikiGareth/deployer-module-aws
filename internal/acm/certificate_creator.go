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

func (cc *certificateCreator) Loc() *errorsink.Location {
	return cc.loc
}

func (cc *certificateCreator) ShortDescription() string {
	return "aws.CertificateManager.Certificate[" + cc.name + "]"
}

func (cc *certificateCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.CertificateManager.Certificate[")
	iw.AttrsWhere(cc)
	iw.TextAttr("named", cc.name)
	iw.EndAttrs()
}

func (cc *certificateCreator) CoinId() corebottom.CoinId {
	return cc.coin
}

func (cc *certificateCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	ae := cc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := ae.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	cc.client = awsEnv.ACMClient()
	cc.route53 = awsEnv.Route53Client()

	certs := cc.findCertificatesFor(cc.name)
	if len(certs) == 0 {
		log.Printf("there were no certs found for %s\n", cc.name)
		pres.NotFound()
	} else {
		model := NewCertificateModel(cc.loc, cc.coin)
		model.name = cc.name

		// log.Printf("found %d certs for %s\n", len(certs), acmc.name)
		model.arn = certs[0]

		// acmc.describeCertificate(acmc.arn)
		// acmc.tools.Storage.Bind(acmc.coin, model)
		pres.Present(model)
	}
}

func (cc *certificateCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	model := NewCertificateModel(cc.loc, cc.coin)
	for k, p := range cc.props {
		v := cc.tools.Storage.Eval(p)
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
		case "ValidationProvider":
			meth, ok := utils.AsStringer(v)
			if !ok {
				log.Fatalf("ValidationProvider must be a string")
				return
			}
			model.validationProvider = meth
		default:
			log.Fatalf("certificate coin does not support a parameter %s\n", k.Id())
		}
	}
	pres.Present(model)
}

func (cc *certificateCreator) UpdateReality() {
	found := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_INITIAL_MODE)

	desired := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_DESIRED_MODE).(*certificateModel)

	vm := types.ValidationMethod(desired.validationMethod.String())
	if vm == "" {
		vm = types.ValidationMethodDns
	}
	vp := desired.validationProvider.String()
	var dnsAsserter func(string, string, string) error
	if vp == "" || vp == "Route53" {
		dnsAsserter = func(zone, key, value string) error {
			return cc.insertCheckRecords(desired, zone, key, value)
		}
	} else {
		tmp := cc.tools.Recall.Find("dns-asserter", vp)
		if tmp == nil {
			panic("no dns-asserter for " + vp + " was found")
		}
		var ok bool
		dnsAsserter, ok = tmp.(func(string, string, string) error)
		if !ok {
			panic(vp + " was not a dns-asserter")
		}
	}

	var created *certificateModel
	if found != nil {
		foundCert := found.(*certificateModel)
		log.Printf("certificate %s already existed for %s\n", foundCert.arn, foundCert.name)
		cc.tools.Storage.Adopt(cc.coin, foundCert)
		created = foundCert
	} else {
		created = NewCertificateModel(desired.loc, cc.coin)
		created.name = desired.name
		created.hzid = desired.hzid

		input := acm.RequestCertificateInput{DomainName: &cc.name, ValidationMethod: vm}
		if len(desired.sans) > 0 {
			input.SubjectAlternativeNames = desired.sans
		}
		req, err := cc.client.RequestCertificate(context.TODO(), &input)
		if err != nil {
			log.Printf("failed to request cert %s: %v\n", cc.name, err)
		}
		log.Printf("requested cert for %s: %s\n", cc.name, *req.CertificateArn)
		created.arn = *req.CertificateArn
	}

	// Either way, may sure it is validated ...
	utils.ExponentialBackoff(func() bool {
		return cc.tryToValidateCert(created.arn, dnsAsserter)
	})

	if found == nil {
		cc.tools.Storage.Bind(cc.coin, created)
	}
}

func (cc *certificateCreator) TearDown() {
	tmp := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp == nil {
		log.Printf("no certificate existed for %s\n", cc.name)
		return
	}

	found := tmp.(*certificateModel)
	log.Printf("you have asked to tear down certificate for %s (arn: %s) with mode %s\n", found.name, found.arn, cc.teardown.Mode())
	switch cc.teardown.Mode() {
	case "preserve":
		log.Printf("not deleting certificate %s because teardown mode is 'preserve'", found.name)
	case "delete":
		log.Printf("deleting certificate for %s with teardown mode 'delete'", found.name)
		DeleteCertificate(cc.client, found.arn)
	default:
		log.Printf("cannot handle teardown mode '%s' for bucket %s", cc.teardown.Mode(), found.name)
	}
}

func (cc *certificateCreator) findCertificatesFor(name string) []string {
	ret := make([]string, 0)
	// TODO: need a loop on "NextToken"
	certs, err := cc.client.ListCertificates(context.TODO(), &acm.ListCertificatesInput{})
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

func (cc *certificateCreator) DescribeCertificate(arn string) {
	cert, err := cc.client.DescribeCertificate(context.TODO(), &acm.DescribeCertificateInput{CertificateArn: &arn})
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

func (cc *certificateCreator) tryToValidateCert(arn string, asserter func(string, string, string) error) bool {
	cert, err := cc.client.DescribeCertificate(context.TODO(), &acm.DescribeCertificateInput{CertificateArn: &arn})
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
		// dns := make(map[string]string, 0)
		for _, x := range cert.Certificate.DomainValidationOptions {
			if x.ResourceRecord != nil && x.ResourceRecord.Name != nil && x.ResourceRecord.Value != nil {
				log.Printf("need %s => %s\n", *x.ResourceRecord.Name, *x.ResourceRecord.Value)
				// dns[*x.ResourceRecord.Name] = *x.ResourceRecord.Value
				err := asserter(cc.name, *x.ResourceRecord.Name, *x.ResourceRecord.Value)
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

func (cc *certificateCreator) insertCheckRecords(model *certificateModel, _, key, value string) error {
	rrs, err := cc.route53.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{HostedZoneId: &model.hzid})
	if err != nil {
		return err
	}
	for _, r := range rrs.ResourceRecordSets {
		if r.Type == "CNAME" {
			if *r.Name == key && *r.ResourceRecords[0].Value == value {
				log.Printf("already have %s %v\n", *r.Name, *r.ResourceRecords[0].Value)
				return nil
			}
		}
	}
	log.Printf("creating %s to %s\n", key, value)
	var ttl int64 = 300
	changes := r53types.ResourceRecordSet{Name: &key, Type: "CNAME", TTL: &ttl, ResourceRecords: []r53types.ResourceRecord{{Value: &value}}}
	cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "CREATE", ResourceRecordSet: &changes}}}
	_, err = cc.route53.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &model.hzid, ChangeBatch: &cb})
	return err
}

func (cc *certificateCreator) String() string {
	return fmt.Sprintf("CreateCert[%s]", cc.name)
}

func DeleteCertificate(client *acm.Client, arn string) {
	_, err := client.DeleteCertificate(context.TODO(), &acm.DeleteCertificateInput{CertificateArn: &arn})
	if err != nil {
		panic(err)
	}
}

var _ corebottom.Ensurable = &certificateCreator{}

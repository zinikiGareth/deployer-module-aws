package acm

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/acm/types"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	r53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
	myroute53 "ziniki.org/deployer/modules/aws/internal/route53"
)

type certificateCreator struct {
	tools *pluggable.Tools

	loc              *errorsink.Location
	name             string
	validationMethod types.ValidationMethod
	teardown         pluggable.TearDown

	client        *acm.Client
	route53       *route53.Client
	alreadyExists bool
	hzid          string
	arn           string
	props         map[pluggable.Identifier]pluggable.Expr
	// cloud *BucketCloud
}

func (acm *certificateCreator) Loc() *errorsink.Location {
	return acm.loc
}

func (acm *certificateCreator) ShortDescription() string {
	return "aws.CertificateManager.Certificate[" + acm.name + "]"
}

func (acm *certificateCreator) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.CertificateManager.Certificate[")
	iw.AttrsWhere(acm)
	iw.TextAttr("named", acm.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (acmc *certificateCreator) BuildModel(pres pluggable.ValuePresenter) {
	domainExpr := find(acmc.props, "Domain")
	if domainExpr == nil {
		log.Fatalf("must specify a domain instance to create a certificate")
	}
	domainObj := domainExpr.Eval(acmc.tools.Storage)
	// log.Printf("have domain obj: %T %p %v\n", domainObj, domainObj, domainObj)
	domain, ok := domainObj.(myroute53.ExportedDomain)
	if !ok {
		log.Fatalf("Domain did not point to a domain instance")
	}
	acmc.hzid = domain.HostedZoneId()
	if acmc.validationMethod == "" {
		acmc.validationMethod = types.ValidationMethodDns
	}
	eq := acmc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	acmc.client = awsEnv.ACMClient()
	acmc.route53 = awsEnv.Route53Client()

	certs := acmc.findCertificatesFor(acmc.name)
	if len(certs) == 0 {
		log.Printf("there were no certs found for %s\n", acmc.name)
	} else {
		// log.Printf("found %d certs for %s\n", len(certs), acmc.name)
		acmc.alreadyExists = true
		acmc.arn = certs[0]
		// acmc.describeCertificate(acmc.arn)
	}

	// TODO: do we need to capture something here?
	pres.Present(acmc)
}

func (acmc *certificateCreator) UpdateReality() {
	if acmc.alreadyExists {
		log.Printf("certificate %s already existed for %s\n", acmc.arn, acmc.name)
		return
	}

	req, err := acmc.client.RequestCertificate(context.TODO(), &acm.RequestCertificateInput{DomainName: &acmc.name, ValidationMethod: acmc.validationMethod})
	if err != nil {
		log.Printf("failed to request cert %s: %v\n", acmc.name, err)
	}
	log.Printf("requested cert for %s: %v\n", acmc.name, req)
	acmc.arn = *req.CertificateArn

	// Check if we still need to validate it ...

	// acmc.describeCertificate(*req.CertificateArn)
	var waitFor time.Duration = 1
	for {
		log.Printf("sleeping for %ds\n", waitFor)
		time.Sleep(waitFor * time.Second)
		if acmc.tryToValidateCert(acmc.arn) {
			break
		}
		waitFor = min(2*waitFor, 60)
		fmt.Printf("still pending validation; wait another %ds\n", waitFor)
	}
}

func (acm *certificateCreator) TearDown() {
	if !acm.alreadyExists {
		log.Printf("no certificate existed for %s\n", acm.name)
		return
	}
	log.Printf("you have asked to tear down certificate for %s (arn: %s) with mode %s\n", acm.name, acm.arn, acm.teardown.Mode())
	switch acm.teardown.Mode() {
	case "preserve":
		log.Printf("not deleting certificate %s because teardown mode is 'preserve'", acm.name)
	case "delete":
		log.Printf("deleting certificate for %s with teardown mode 'delete'", acm.name)
		DeleteCertificate(acm.client, acm.arn)
	default:
		log.Printf("cannot handle teardown mode '%s' for bucket %s", acm.teardown.Mode(), acm.name)
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
	if cert.Certificate.Status == types.CertificateStatusFailed {
		fmt.Printf("failed: %s\n", cert.Certificate.FailureReason)
	} else if cert.Certificate.Status == types.CertificateStatusIssued {
		fmt.Printf("until: %s\n", *cert.Certificate.NotAfter)
	} else if cert.Certificate.Status == types.CertificateStatusPendingValidation {
		fmt.Printf("pending validation\n")
	}
}

func (acmc *certificateCreator) tryToValidateCert(arn string) bool {
	cert, err := acmc.client.DescribeCertificate(context.TODO(), &acm.DescribeCertificateInput{CertificateArn: &arn})
	if err != nil {
		log.Fatalf("Failed to describe certificate %s: %v\n", arn, err)
	}

	if cert.Certificate.Status == types.CertificateStatusFailed {
		fmt.Printf("failed: %s\n", cert.Certificate.FailureReason)
		return true
	} else if cert.Certificate.Status == types.CertificateStatusIssued {
		fmt.Printf("certificate issued until: %s\n", *cert.Certificate.NotAfter)
		return true
	} else if cert.Certificate.Status == types.CertificateStatusPendingValidation {
		dns := make(map[string]string, 0)
		for _, x := range cert.Certificate.DomainValidationOptions {
			if x.ResourceRecord != nil && x.ResourceRecord.Name != nil && x.ResourceRecord.Value != nil {
				fmt.Printf("need %s => %s\n", *x.ResourceRecord.Name, *x.ResourceRecord.Value)
				dns[*x.ResourceRecord.Name] = *x.ResourceRecord.Value
			}
		}
		rrs, err := acmc.route53.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{HostedZoneId: &acmc.hzid})
		if err != nil {
			panic(err)
		}
		for _, r := range rrs.ResourceRecordSets {
			if r.Type == "CNAME" {
				fmt.Printf("already have %s %v\n", *r.Name, *r.ResourceRecords[0].Value)
				dns[*r.Name] = ""
			}
		}
		for k, v := range dns {
			if v != "" {
				fmt.Printf("creating %s to %s\n", k, v)
				var ttl int64 = 300
				changes := r53types.ResourceRecordSet{Name: &k, Type: "CNAME", TTL: &ttl, ResourceRecords: []r53types.ResourceRecord{{Value: &v}}}
				cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "CREATE", ResourceRecordSet: &changes}}}
				_, err := acmc.route53.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &acmc.hzid, ChangeBatch: &cb})
				if err != nil {
					panic(err)
				}
			}
		}

		return false
	} else {
		panic("what is this? " + cert.Certificate.Status)
	}
}

func (acm *certificateCreator) ObtainMethod(name string) pluggable.Method {
	switch name {
	case "arn":
		return &arnMethod{}
	}
	return nil
}

func (acm *certificateCreator) String() string {
	return fmt.Sprintf("EnsureBucket[%s:%s]", "" /* eacm.env.Region */, acm.name)
}

type arnMethod struct {
}

func (a *arnMethod) Invoke(s pluggable.RuntimeStorage, on pluggable.Expr, args []pluggable.Expr) any {
	e := on.Eval(s)
	cc, ok := e.(*certificateCreator)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a certificate, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if cc.alreadyExists {
		return cc.arn
	} else {
		return &DeferReadingArn{cc: cc}
	}
}

type DeferReadingArn struct {
	cc *certificateCreator
}

func (d *DeferReadingArn) String() string {
	if d.cc.arn == "" {
		panic("arn is still not set")
	}
	return d.cc.arn
}

func find(props map[pluggable.Identifier]pluggable.Expr, key string) pluggable.Expr {
	for id, ex := range props {
		if id.Id() == key {
			return ex
		}
	}
	return nil
}

func DeleteCertificate(client *acm.Client, arn string) {
	_, err := client.DeleteCertificate(context.TODO(), &acm.DeleteCertificateInput{CertificateArn: &arn})
	if err != nil {
		panic(err)
	}
	log.Printf("I think the certificate was deleted because no error was reported")
}

var _ pluggable.HasMethods = &certificateCreator{}

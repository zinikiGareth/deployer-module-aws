package acm

import (
	"fmt"
	"log"

	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type certificateFinder struct {
	tools *pluggable.Tools

	loc  *errorsink.Location
	name string
}

func (acm *certificateFinder) Loc() *errorsink.Location {
	return acm.loc
}

func (acm *certificateFinder) ShortDescription() string {
	return "aws.CertificateManager.Certificate[" + acm.name + "]"
}

func (acm *certificateFinder) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.CertificateManager.Certificate[")
	iw.AttrsWhere(acm)
	iw.TextAttr("named", acm.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (acm *certificateFinder) Prepare(pres pluggable.ValuePresenter) {
	eq := acm.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	/*
		acm.client = awsEnv.Route53DomainsClient()
		_, err := acm.client.GetDomainDetail(context.TODO(), &route53domains.GetDomainDetailInput{DomainName: &acm.name})
		if err != nil {
			var api smithy.APIError
			if e.As(err, &api) {
				if api.ErrorCode() == "NotFound" || strings.Contains(api.ErrorMessage(), "not found in account") {
					acm.tools.Reporter.At(acm.loc.Line)
					acm.tools.Reporter.Reportf(acm.loc.Offset, "domain does not belong to this account: %s\n", acm.name)
				} else {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		}
	*/
	log.Printf("%v\n", awsEnv)
	pres.Present(acm)
}

func (acm *certificateFinder) Execute() {
}

func (acm *certificateFinder) String() string {
	return fmt.Sprintf("FindDomainName[%s]", acm.name)
}

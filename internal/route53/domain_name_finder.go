package route53

import (
	"context"
	e "errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type domainNameFinder struct {
	tools *pluggable.Tools

	loc    *errorsink.Location
	name   string
	client *route53domains.Client
}

func (dnf *domainNameFinder) Loc() *errorsink.Location {
	return dnf.loc
}

func (dnf *domainNameFinder) ShortDescription() string {
	return "aws.route53.DomainName[" + dnf.name + "]"
}

func (dnf *domainNameFinder) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.s3.DomainName[")
	iw.AttrsWhere(dnf)
	iw.TextAttr("named", dnf.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (dnf *domainNameFinder) Prepare(pres pluggable.ValuePresenter) {
	eq := dnf.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	dnf.client = awsEnv.Route53DomainsClient()
	_, err := dnf.client.GetDomainDetail(context.TODO(), &route53domains.GetDomainDetailInput{DomainName: &dnf.name})
	if err != nil {
		var api smithy.APIError
		if e.As(err, &api) {
			if api.ErrorCode() == "NotFound" || strings.Contains(api.ErrorMessage(), "not found in account") {
				dnf.tools.Reporter.At(dnf.loc.Line)
				dnf.tools.Reporter.Reportf(dnf.loc.Offset, "domain does not belong to this account: %s\n", dnf.name)
			} else {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
	// log.Printf("%v\n", detail)
	// TODO: Do we need to capture this?  Or just use the name?
	pres.Present(dnf)
}

func (dnf *domainNameFinder) Execute() {
}

func (dnf *domainNameFinder) String() string {
	return fmt.Sprintf("FindDomainName[%s]", dnf.name)
}

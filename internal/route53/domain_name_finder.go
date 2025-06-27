package route53

import (
	"context"
	e "errors"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type ExportedDomain interface {
	HostedZoneId() string
}

type domainNameFinder struct {
	tools *corebottom.Tools

	loc           *errorsink.Location
	name          string
	route53Client *route53.Client
	domainsClient *route53domains.Client
}

func (dnf *domainNameFinder) Loc() *errorsink.Location {
	return dnf.loc
}

func (dnf *domainNameFinder) ShortDescription() string {
	return "aws.route53.DomainName[" + dnf.name + "]"
}

func (dnf *domainNameFinder) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.s3.DomainName[")
	iw.AttrsWhere(dnf)
	iw.TextAttr("named", dnf.name)
	iw.EndAttrs()
}

func (dnf *domainNameFinder) DetermineInitialState(pres driverbottom.ValuePresenter) {
	eq := dnf.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	dnf.domainsClient = awsEnv.Route53DomainsClient()
	dnf.route53Client = awsEnv.Route53Client()
	detail, err := dnf.domainsClient.GetDomainDetail(context.TODO(), &route53domains.GetDomainDetailInput{DomainName: &dnf.name})
	if err != nil {
		var api smithy.APIError
		if e.As(err, &api) {
			if api.ErrorCode() == "NotFound" || strings.Contains(api.ErrorMessage(), "not found in account") {
				dnf.tools.Reporter.ReportAtf(dnf.loc, "domain does not belong to this account: %s\n", dnf.name)
				return
			} else {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}

	zones, err := dnf.route53Client.ListHostedZones(context.TODO(), &route53.ListHostedZonesInput{})
	if err != nil {
		panic(err)
	}
	var hzid string
	for _, z := range zones.HostedZones {
		if *z.Name == dnf.name+"." {
			hzid = strings.Replace(*z.Id, "/hostedzone/", "", 1)
			log.Printf("found zone %s: %s\n", hzid, *z.Name)
		}
	}
	if hzid == "" {
		log.Fatalf("No hosted zone found for " + dnf.name)
	}
	model := CreateDomainModel(dnf.loc, detail, hzid)
	pres.Present(model)
}

func (dnf *domainNameFinder) String() string {
	return fmt.Sprintf("FindDomainName[%s]", dnf.name)
}

var _ corebottom.FindCoin = &domainNameFinder{}

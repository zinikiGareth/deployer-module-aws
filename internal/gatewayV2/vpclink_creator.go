package gatewayV2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type vpcLinkCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *apigatewayv2.Client
}

func (ic *vpcLinkCreator) Loc() *errorsink.Location {
	return ic.loc
}

func (ic *vpcLinkCreator) ShortDescription() string {
	return "aws.gateway.VPCLink[" + ic.name + "]"
}

func (ic *vpcLinkCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.gateway.VPCLink[")
	iw.AttrsWhere(ic)
	iw.TextAttr("name", ic.name)
	iw.EndAttrs()
}

func (ic *vpcLinkCreator) CoinId() corebottom.CoinId {
	return ic.coin
}

func (ic *vpcLinkCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := ic.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	ic.client = awsEnv.ApiGatewayV2Client()

	var nextTok *string
	var wanted *types.VpcLink
outer:
	for {
		curr, err := ic.client.GetVpcLinks(context.TODO(), &apigatewayv2.GetVpcLinksInput{NextToken: nextTok})
		if err != nil {
			log.Fatalf("could not recover integration list: %v\n", err)
		}
		for _, vl := range curr.Items {
			if vl.Name != nil && *vl.Name == ic.name {
				wanted = &vl
				break outer
			}
		}
		if curr.NextToken == nil {
			break
		}
		nextTok = curr.NextToken
	}
	if wanted == nil {
		log.Printf("did not find VpcLink called %s\n", ic.name)
		pres.NotFound()
		return
	}

	log.Printf("found VpcLink %s\n", *wanted.VpcLinkId)

	model := &VPCLinkAWSModel{link: wanted}
	pres.Present(model)
}

func (ic *vpcLinkCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var subnets driverbottom.Expr
	var groups driverbottom.Expr
	for p, v := range ic.props {
		switch p.Id() {
		case "Subnets":
			subnets = v
		case "SecurityGroups":
			groups = v
		default:
			ic.tools.Reporter.ReportAtf(p.Loc(), "invalid property for vpc link: %s", p.Id())
		}
	}
	if subnets == nil {
		ic.tools.Reporter.ReportAtf(ic.loc, "no subnets specified for VPCLink")
		return
	}
	if groups == nil {
		ic.tools.Reporter.ReportAtf(ic.loc, "no security groups specified for VPCLink")
		return
	}

	model := &VPCLinkModel{name: ic.name, loc: ic.loc, coin: ic.coin, subnets: subnets, groups: groups}
	pres.Present(model)
}

func (ic *vpcLinkCreator) UpdateReality() {
	tmp := ic.tools.Storage.GetCoin(ic.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := ic.tools.Storage.GetCoin(ic.coin, corebottom.DETERMINE_DESIRED_MODE).(*VPCLinkModel)
	created := &VPCLinkAWSModel{}

	subnets := desired.subnets.Eval(ic.tools.Storage).([]string)
	groups := desired.groups.Eval(ic.tools.Storage).([]string)
	if tmp != nil {
		found := tmp.(*VPCLinkAWSModel)
		created.link = found.link

		log.Printf("vpclink already existed called %s\n", ic.name)
		log.Printf("not handling diffs yet; just copying ...")
		ic.tools.Storage.Bind(ic.coin, created)
	} else {
		input := &apigatewayv2.CreateVpcLinkInput{Name: &ic.name, SubnetIds: subnets, SecurityGroupIds: groups}
		out, err := ic.client.CreateVpcLink(context.TODO(), input)
		if err != nil {
			log.Fatalf("failed to create vpc link %s: %v\n", ic.name, err)
		}
		created.link = &types.VpcLink{Name: &ic.name, VpcLinkId: out.VpcLinkId}
		ic.tools.Storage.Bind(ic.coin, created)
	}
	utils.ExponentialBackoff(func() bool {
		stat, err := ic.client.GetVpcLink(context.TODO(), &apigatewayv2.GetVpcLinkInput{VpcLinkId: created.link.VpcLinkId})
		if err != nil {
			panic(err)
		}
		if stat.VpcLinkStatus == types.VpcLinkStatusAvailable {
			return true
		}
		log.Printf("waiting for VPC Link to be available, stat = %v\n", stat.VpcLinkStatus)
		return false
	})

	log.Printf("created vpc link %s\n", *created.link.VpcLinkId)
}

func (ic *vpcLinkCreator) TearDown() {
	tmp := ic.tools.Storage.GetCoin(ic.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*VPCLinkAWSModel)
		log.Printf("you have asked to tear down vpc link %s with mode %s\n", ic.name, ic.teardown.Mode())

		_, err := ic.client.DeleteVpcLink(context.TODO(), &apigatewayv2.DeleteVpcLinkInput{VpcLinkId: found.link.VpcLinkId})
		if err != nil {
			log.Fatalf("failed to delete vpc link %s: %v\n", ic.name, err)
		}
	} else {
		log.Printf("no vpc link existed called %s\n", ic.name)
	}

}

var _ corebottom.Ensurable = &vpcLinkCreator{}

package vpc

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type vpcFinder struct {
	tools *corebottom.Tools

	loc       *errorsink.Location
	name      string
	coin      corebottom.CoinId
	vpcClient *ec2.Client
}

func (vpc *vpcFinder) Loc() *errorsink.Location {
	return vpc.loc
}

func (vpc *vpcFinder) ShortDescription() string {
	return "aws.VPC.VPC[" + vpc.name + "]"
}

func (vpc *vpcFinder) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.VPC.VPC[")
	iw.AttrsWhere(vpc)
	iw.TextAttr("name", vpc.name)
	iw.EndAttrs()
}

func (vpc *vpcFinder) CoinId() corebottom.CoinId {
	return vpc.coin
}

func (vpc *vpcFinder) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := vpc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	vpc.vpcClient = awsEnv.EC2Client()

	var nextTok *string
	var wanted *types.Vpc
outer:
	for {
		curr, err := vpc.vpcClient.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{NextToken: nextTok})
		if err != nil {
			log.Fatalf("could not recover integration list: %v\n", err)
		}
		for _, intg := range curr.Vpcs {
			for _, t := range intg.Tags {
				if t.Key != nil && *t.Key == "Name" && t.Value != nil && *t.Value == vpc.name {
					wanted = &intg
					break outer
				}
			}
		}
		if curr.NextToken == nil {
			log.Printf("did not find VPC called %s\n", vpc.name)
			pres.NotFound()
			return
		}
	}

	model := &vpcAWSModel{loc: vpc.loc, vpc: wanted}
	pres.Present(model)
}

func (vpc *vpcFinder) String() string {
	return fmt.Sprintf("FindVPC[%s]", vpc.name)
}

var _ corebottom.FindCoin = &vpcFinder{}

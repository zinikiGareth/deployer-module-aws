package vpc

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
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

func (vf *vpcFinder) Loc() *errorsink.Location {
	return vf.loc
}

func (vf *vpcFinder) ShortDescription() string {
	return "aws.VPC.VPC[" + vf.name + "]"
}

func (vf *vpcFinder) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.VPC.VPC[")
	iw.AttrsWhere(vf)
	iw.TextAttr("name", vf.name)
	iw.EndAttrs()
}

func (vf *vpcFinder) CoinId() corebottom.CoinId {
	return vf.coin
}

func (vf *vpcFinder) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := vf.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	vf.vpcClient = awsEnv.EC2Client()

	var nextTok *string
	var wanted *types.Vpc
outer:
	for {
		curr, err := vf.vpcClient.DescribeVpcs(context.TODO(), &ec2.DescribeVpcsInput{NextToken: nextTok})
		if err != nil {
			log.Fatalf("could not recover integration list: %v\n", err)
		}
		for _, intg := range curr.Vpcs {
			for _, t := range intg.Tags {
				if t.Key != nil && *t.Key == "Name" && t.Value != nil && *t.Value == vf.name {
					wanted = &intg
					break outer
				}
			}
		}
		if curr.NextToken == nil {
			break
		}
		nextTok = curr.NextToken
	}
	if wanted == nil {
		log.Printf("did not find VPC called %s\n", vf.name)
		pres.NotFound()
		return
	}
	log.Printf("have vpc %s\n", *wanted.VpcId)

	// find its subnets
	nextTok = nil
	var subnetIds []string
	for {
		dsi := &ec2.DescribeSubnetsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{*wanted.VpcId},
				},
			},
			NextToken: nextTok,
		}
		curr, err := vf.vpcClient.DescribeSubnets(context.TODO(), dsi)
		if err != nil {
			panic(err)
		}
		for _, sn := range curr.Subnets {
			subnetIds = append(subnetIds, *sn.SubnetId)
		}
		if curr.NextToken == nil {
			break
		}
		nextTok = curr.NextToken
	}
	for _, sn := range subnetIds {
		log.Printf("have subnet %s", sn)
	}

	// find its subnets
	nextTok = nil
	var secgroups []string
	for {
		dsi := &ec2.DescribeSecurityGroupsInput{
			Filters: []types.Filter{
				{
					Name:   aws.String("vpc-id"),
					Values: []string{*wanted.VpcId},
				},
			},
			NextToken: nextTok,
		}
		curr, err := vf.vpcClient.DescribeSecurityGroups(context.TODO(), dsi)
		if err != nil {
			panic(err)
		}
		for _, sg := range curr.SecurityGroups {
			secgroups = append(secgroups, *sg.GroupId)
		}
		if curr.NextToken == nil {
			break
		}
		nextTok = curr.NextToken
	}
	for _, sn := range secgroups {
		log.Printf("have security group %s", sn)
	}

	model := &vpcAWSModel{loc: vf.loc, vpc: wanted, subnets: subnetIds, securityGroups: secgroups}
	pres.Present(model)
}

func (vf *vpcFinder) String() string {
	return fmt.Sprintf("FindVPC[%s]", vf.name)
}

var _ corebottom.FindCoin = &vpcFinder{}

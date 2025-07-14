package neptune

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type subnetCreator struct {
	tools *corebottom.Tools

	loc  *errorsink.Location
	name string
	coin corebottom.CoinId
	// teardown corebottom.TearDown
	// props    map[driverbottom.Identifier]driverbottom.Expr

	client *neptune.Client
}

func (cc *subnetCreator) Loc() *errorsink.Location {
	return cc.loc
}

func (cc *subnetCreator) ShortDescription() string {
	return "aws.Neptune.SubnetGroup[" + cc.name + "]"
}

func (cc *subnetCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.Neptune.SubnetGroup[")
	iw.AttrsWhere(cc)
	iw.TextAttr("named", cc.name)
	iw.EndAttrs()
}

func (cc *subnetCreator) CoinId() corebottom.CoinId {
	return cc.coin
}

func (cc *subnetCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	log.Printf("want to find neptune cluster %s", cc.name)
	ae := cc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := ae.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	cc.client = awsEnv.NeptuneClient()

	model := cc.findSubnetsNamed(cc.name)

	if model == nil {
		log.Printf("subnet %s not found\n", cc.name)
		pres.NotFound()
	} else {
		log.Printf("subnet found for %s\n", cc.name)
		pres.Present(model)
	}
}

func (cc *subnetCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	panic("should not come in here (yet) - i.e. implement this")
}

func (cc *subnetCreator) UpdateReality() {
	panic("should not come in here (yet) - i.e. implement this")
}

func (cc *subnetCreator) TearDown() {
	panic("should not come in here (yet) - i.e. implement this")
}

func (cc *subnetCreator) findSubnetsNamed(name string) *subnetModel {
	clusters, err := cc.client.DescribeDBSubnetGroups(context.TODO(), &neptune.DescribeDBSubnetGroupsInput{DBSubnetGroupName: &name})
	if !clusterExists(err) {
		return nil
	}
	if len(clusters.DBSubnetGroups) == 0 {
		return nil
	} else if len(clusters.DBSubnetGroups) > 1 {
		log.Fatalf("More than one subnet group called %s found", name)
		panic("multiple subnet groups")
	} else {
		c1 := clusters.DBSubnetGroups[0]
		return NewSubnetGroupModel(cc.loc, cc.coin, *c1.DBSubnetGroupName, *c1.DBSubnetGroupArn)
	}
}

func (cc *subnetCreator) String() string {
	return fmt.Sprintf("subnetCreator[%s]", cc.name)
}

var _ corebottom.Ensurable = &subnetCreator{}

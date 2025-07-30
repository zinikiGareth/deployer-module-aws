package lambda

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type lambdaVersioner struct {
	tools *corebottom.Tools
	coin  corebottom.CoinId
	props map[driverbottom.Identifier]driverbottom.Expr

	client *lambda.Client
}

func (v *lambdaVersioner) Loc() *errorsink.Location {
	panic("unimplemented")
}

func (v *lambdaVersioner) ShortDescription() string {
	panic("unimplemented")
}

func (v *lambdaVersioner) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

func (v *lambdaVersioner) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	ret := driverbottom.MAY_BE_BOUND
	for _, e := range v.props {
		ret = ret.Merge(e.Resolve(r))
	}
	return ret
}

func (v *lambdaVersioner) DetermineInitialState(pres corebottom.ValuePresenter) {
	log.Printf("I can't really say what should happen here yet, but maybe nothing")
	eq := v.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	v.client = awsEnv.LambdaClient()
	// I think we probably want to find the current list of aliases and versions
}

func (v *lambdaVersioner) DetermineDesiredState(pres corebottom.ValuePresenter) {
	pv := utils.F64AsNumber(0)
	if utils.HasProp(v.props, "PublishVersion") {
		pv = v.tools.Storage.EvalAsNumber(utils.FindProp(v.props, nil, "PublishVersion"))
	}
	alias, _ := utils.AsStringer("")
	if utils.HasProp(v.props, "Alias") {
		prop := utils.FindProp(v.props, nil, "Alias")
		s, ok := v.tools.Storage.EvalAsStringer(prop)
		if !ok {
			v.tools.Reporter.ReportAtf(prop.Loc(), "Alias must be a string, not %T", prop)
			return
		}
		alias = s
	}
	prop := utils.FindProp(v.props, nil, "Name")
	name, ok := v.tools.Storage.EvalAsStringer(prop)
	if !ok {
		v.tools.Reporter.ReportAtf(prop.Loc(), "Name must be a string, not %T", prop)
		return
	}

	pres.Present(&publishVersionModel{publish: pv, asAlias: alias, name: name})
}

func (v *lambdaVersioner) ShouldDestroy() bool {
	panic("unimplemented")
}

func (v *lambdaVersioner) UpdateReality() {
	desired := v.tools.Storage.GetCoin(v.coin, corebottom.DETERMINE_DESIRED_MODE).(*publishVersionModel)
	if desired.publish.F64() != 0 {
		name := desired.name.String()
		out, err := v.client.PublishVersion(context.TODO(), &lambda.PublishVersionInput{FunctionName: &name})
		if err != nil {
			log.Fatalf("failed to publish new version of %s: %v", name, err)
		}
		log.Printf("publish returned %s\n", *out.Version)
	}
}

func (v *lambdaVersioner) TearDown() {
	log.Printf("What would it mean to tear down a version?  I think deleting old versions is a separate op")
}

var _ corebottom.RealityShifter = &lambdaVersioner{}

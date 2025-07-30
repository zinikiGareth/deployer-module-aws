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
	eq := v.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	v.client = awsEnv.LambdaClient()

	alias := ""
	if utils.HasProp(v.props, "Alias") {
		prop := utils.FindProp(v.props, nil, "Alias")
		s, ok := v.tools.Storage.EvalAsStringer(prop)
		if !ok {
			v.tools.Reporter.ReportAtf(prop.Loc(), "Alias must be a string, not %T", prop)
			return
		}
		alias = s.String()
	}

	prop := utils.FindProp(v.props, nil, "Name")
	name, ok := v.tools.Storage.EvalAsStringer(prop)
	if !ok {
		v.tools.Reporter.ReportAtf(prop.Loc(), "Name must be a string, not %T", prop)
		return
	}
	fname := name.String()

	if fname == "" { // if the function name is nil, that's probably because it hasn't been created (or has been destroyed), so we don't stand a chance of finding it ...
		pres.NotFound()
		return
	}

	out, err := v.client.GetAlias(context.TODO(), &lambda.GetAliasInput{FunctionName: &fname, Name: &alias})
	if err != nil {
		if !lambdaExists(err) {
			log.Printf("no alias found for %s:%s\n", fname, name)
			pres.NotFound()
			return
		}
		log.Fatalf("failed to get alias %s:%s: %v\n", fname, alias, err)
	}

	log.Printf("found alias version %s for %s\n", *out.FunctionVersion, *out.Name)
	model := &publishVersionAWS{functionName: fname, aliasName: alias, aliasVersion: *out.FunctionVersion, aliasRevId: *out.RevisionId}
	pres.Present(model)
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
	tmp := v.tools.Storage.GetCoin(v.coin, corebottom.DETERMINE_INITIAL_MODE)
	var found *publishVersionAWS
	if tmp != nil {
		found = tmp.(*publishVersionAWS)
		log.Printf("have found version %p %T %v\n", found, found, found)
	}
	desired := v.tools.Storage.GetCoin(v.coin, corebottom.DETERMINE_DESIRED_MODE).(*publishVersionModel)
	created := &publishVersionAWS{}
	name := desired.name.String()
	if desired.publish.F64() != 0 {
		out, err := v.client.PublishVersion(context.TODO(), &lambda.PublishVersionInput{FunctionName: &name})
		if err != nil {
			log.Fatalf("failed to publish new version of %s: %v", name, err)
		}
		log.Printf("published version %s of %s\n", *out.Version, name)
		created.publishedVersion = *out.Version
	}
	if alias := desired.asAlias.String(); alias != "" {
		if found == nil || found.aliasVersion == "" {
			out, err := v.client.CreateAlias(context.TODO(), &lambda.CreateAliasInput{FunctionName: &name, Name: &alias, FunctionVersion: &created.publishedVersion})
			if err != nil {
				log.Fatalf("failed to create alias %s:%s %v", name, alias, err)
			}
			log.Printf("alias returned %s\n", *out.AliasArn)
			created.aliasVersion = *out.FunctionVersion
			created.aliasRevId = *out.RevisionId
		}
	}
	v.tools.Storage.Bind(v.coin, created)
}

func (v *lambdaVersioner) TearDown() {
	tmp := v.tools.Storage.GetCoin(v.coin, corebottom.DETERMINE_INITIAL_MODE)
	var found *publishVersionAWS
	if tmp != nil {
		found = tmp.(*publishVersionAWS)
		_, err := v.client.DeleteAlias(context.TODO(), &lambda.DeleteAliasInput{FunctionName: &found.functionName, Name: &found.aliasName})
		if err != nil {
			log.Fatalf("failed to create alias %s:%s %v", found.functionName, found.aliasName, err)
		}
		log.Printf("deleted alias %s:%s\n", found.functionName, found.aliasName)
	} else {
		log.Printf("no lambda alias to tear down")
	}
	log.Printf("What would it mean to tear down a version?  I think cleaning up old versions is a separate op")
}

var _ corebottom.RealityShifter = &lambdaVersioner{}

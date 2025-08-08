package lambda

import (
	"context"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type lambdaVersioner struct {
	tools *corebottom.Tools
	loc   *errorsink.Location
	coin  corebottom.CoinId
	props map[driverbottom.Identifier]driverbottom.Expr

	defaultPV float64

	client *lambda.Client
}

func (v *lambdaVersioner) AddAdverb(adverb driverbottom.Adverb, args []driverbottom.Token) driverbottom.Interpreter {
	panic("unimplemented")
}

func (v *lambdaVersioner) AddProperty(name driverbottom.Identifier, expr driverbottom.Expr) {
	v.props[name] = expr
}

func (v *lambdaVersioner) Completed() {
	v.coin = corebottom.CoinId(v.tools.Storage.PendingObjId(v.loc))
}

func (v *lambdaVersioner) Loc() *errorsink.Location {
	return v.loc
}

func (v *lambdaVersioner) ShortDescription() string {
	panic("unimplemented")
}

func (v *lambdaVersioner) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

func (v *lambdaVersioner) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	ret := driverbottom.MAY_BE_BOUND
	v.coin.Resolve(v.tools.Storage)
	for _, e := range v.props {
		ret = ret.Merge(e.Resolve(r))
	}
	return ret
}

func (v *lambdaVersioner) CoinId() corebottom.CoinId {
	return v.coin
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
		pres.NotFound()
		return
	}
	fname := name.String()

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
	pv := utils.F64AsNumber(v.defaultPV)
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
	}
	desired := v.tools.Storage.GetCoin(v.coin, corebottom.DETERMINE_DESIRED_MODE).(*publishVersionModel)
	created := &publishVersionAWS{}
	name := desired.name.String()
	if desired.publish.F64() != 0 {
		utils.ExponentialBackoff(func() bool {
			out, err := v.client.PublishVersion(context.TODO(), &lambda.PublishVersionInput{FunctionName: &name})
			if err != nil {
				if isUpdatingFunction(err) {
					log.Printf("still updating function code; cannot publish yet")
					return false
				}
				log.Fatalf("failed to publish new version of %s: %v", name, err)
			}
			log.Printf("published version %s of %s\n", *out.Version, name)
			created.publishedVersion = *out.Version
			return true
		})
	}
	if alias := desired.asAlias.String(); alias != "" {
		if found == nil || found.aliasVersion == "" {
			out, err := v.client.CreateAlias(context.TODO(), &lambda.CreateAliasInput{FunctionName: &name, Name: &alias, FunctionVersion: &created.publishedVersion})
			if err != nil {
				log.Fatalf("failed to create alias %s:%s %v", name, alias, err)
			}
			log.Printf("alias returned %s for version %s\n", *out.AliasArn, *out.FunctionVersion)
			created.aliasVersion = *out.FunctionVersion
			created.aliasRevId = *out.RevisionId
		} else {
			out, err := v.client.UpdateAlias(context.TODO(), &lambda.UpdateAliasInput{FunctionName: &name, Name: &alias, FunctionVersion: &created.publishedVersion})
			if err != nil {
				log.Fatalf("failed to create alias %s:%s %v", name, alias, err)
			}
			log.Printf("alias returned %s for version %s\n", *out.AliasArn, *out.FunctionVersion)
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

var _ corebottom.CoinProvider = &lambdaVersioner{}
var _ corebottom.RealityShifter = &lambdaVersioner{}

func isUpdatingFunction(err error) bool {
	if err == nil {
		return false
	}
	e1, ok := err.(*smithy.OperationError)
	if ok {
		e2, ok := e1.Err.(*http.ResponseError)
		if ok {
			// log.Printf("%v %s", e2, e2.Err)
			if e2.ResponseError.Response.StatusCode == 409 {
				switch e4 := e2.Err.(type) {
				case *types.ResourceConflictException:
					// log.Printf("error: %T %v %s", e4, e4, e4.ErrorMessage())
					s := e4.ErrorMessage()
					if strings.Contains(s, "An update is in progress") {
						return true
					}
				default:
					log.Printf("not invalid: %T %v", e4, e4)
					panic("what error?")
				}
			}
			log.Fatalf("error: %T %v %T %v", e2.Response.Status, e2.Response.Status, e2.ResponseError.Response.StatusCode, e2.ResponseError.Response.StatusCode)
		}
		log.Fatalf("not response error: %T %v", e1.Err, e1.Err)
	}
	log.Fatalf("creating lambda failed: %T %v", err, err)
	panic("failed")
}

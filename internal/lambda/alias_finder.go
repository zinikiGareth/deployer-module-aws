package lambda

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type aliasFinder struct {
	tools *corebottom.Tools

	loc   *errorsink.Location
	name  string
	coin  corebottom.CoinId
	props map[driverbottom.Identifier]driverbottom.Expr

	client *lambda.Client
}

func (lc *aliasFinder) Loc() *errorsink.Location {
	return lc.loc
}

func (lc *aliasFinder) ShortDescription() string {
	return "aws.Lambda.Alias[" + lc.name + "]"
}

func (lc *aliasFinder) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.Lambda.Alias[")
	iw.AttrsWhere(lc)
	iw.TextAttr("named", lc.name)
	iw.NestedAttr("function", utils.FindProp(lc.props, nil, "FunctionName"))
	iw.EndAttrs()
}

func (lc *aliasFinder) CoinId() corebottom.CoinId {
	return lc.coin
}

func (lc *aliasFinder) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := lc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	lc.client = awsEnv.LambdaClient()

	// TODO: needs proper processing
	fnStr, ok := lc.tools.Storage.EvalAsStringer(utils.FindProp(lc.props, nil, "FunctionName"))
	if !ok {
		panic("not ok to not be able to <find> something")
	}

	req, err := lc.client.GetAlias(context.TODO(), &lambda.GetAliasInput{Name: &lc.name, FunctionName: aws.String(fnStr.String())})
	if err != nil {
		if !lambdaExists(err) {
			pres.NotFound()
			return
		}
		log.Fatalf("error trying to find alias %s: %v\n", lc.name, err)
	}
	if req == nil {
		pres.NotFound()
		return
	}
	model := &LambdaAliasAWSModel{name: lc.name, function: fnStr.String(), arn: *req.AliasArn}
	pres.Present(model)
}

var _ corebottom.FindCoin = &aliasFinder{}

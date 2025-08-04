package gatewayV2

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type stageCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *apigatewayv2.Client
}

func (sc *stageCreator) Loc() *errorsink.Location {
	return sc.loc
}

func (sc *stageCreator) ShortDescription() string {
	return "aws.gateway.Stage[" + sc.name + "]"
}

func (sc *stageCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.gateway.Stage[")
	iw.AttrsWhere(sc)
	iw.TextAttr("name", sc.name)
	iw.EndAttrs()
}

func (sc *stageCreator) CoinId() corebottom.CoinId {
	return sc.coin
}

func (sc *stageCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := sc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	sc.client = awsEnv.ApiGatewayV2Client()

	if !utils.HasProp(sc.props, "Api") {
		pres.NotFound()
		return
	}
	ae := utils.FindProp(sc.props, nil, "Api")
	apiStr, ok := sc.tools.Storage.EvalAsStringer(ae)
	if apiStr == nil {
		// if we can't resolve apiId, we won't be able to find it :-)
		pres.NotFound()
		return
	}
	if !ok {
		panic("not ok")
	}
	apiId := apiStr.String()

	_, err := sc.client.GetStage(context.TODO(), &apigatewayv2.GetStageInput{ApiId: &apiId, StageName: &sc.name})
	if err != nil {
		if !thingExists(err) {
			pres.NotFound()
			return
		}
		log.Fatalf("could not recover stage %s: %v\n", sc.name, err)
	}
	log.Printf("found stage %s\n", sc.name)
	model := &StageAWSModel{name: sc.name}
	pres.Present(model)
}

func (sc *stageCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var api driverbottom.Expr
	for p, v := range sc.props {
		switch p.Id() {
		case "Api":
			api = v
		default:
			sc.tools.Reporter.ReportAtf(p.Loc(), "invalid property for Api stage: %s", p.Id())
		}
	}
	if api == nil {
		sc.tools.Reporter.ReportAtf(sc.loc, "no Api specified for Route")
		return
	}

	apiStr, ok := sc.tools.Storage.EvalAsStringer(api)
	if !ok {
		panic("not ok")
	}

	model := &StageModel{name: sc.name, loc: sc.loc, coin: sc.coin, api: apiStr}
	pres.Present(model)
}

func (sc *stageCreator) UpdateReality() {
	tmp := sc.tools.Storage.GetCoin(sc.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := sc.tools.Storage.GetCoin(sc.coin, corebottom.DETERMINE_DESIRED_MODE).(*StageModel)
	created := &StageAWSModel{}
	apiId := desired.api.String()

	if tmp != nil {
		found := tmp.(*StageAWSModel)
		created.name = found.name

		log.Printf("stage already existed for %s: %s\n", apiId, sc.name)
		log.Printf("not handling diffs yet; just copying ...")
		sc.tools.Storage.Bind(sc.coin, created)
		return
	}

	input := &apigatewayv2.CreateStageInput{ApiId: &apiId, StageName: &sc.name}
	out, err := sc.client.CreateStage(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to create api stage %s: %v\n", sc.name, err)
	}
	log.Printf("created api stage %s\n", *out.StageName)
	created.name = sc.name
	sc.tools.Storage.Bind(sc.coin, created)

}

func (sc *stageCreator) TearDown() {
	tmp := sc.tools.Storage.GetCoin(sc.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := sc.tools.Storage.GetCoin(sc.coin, corebottom.DETERMINE_DESIRED_MODE).(*StageModel)

	apiId := desired.api.String()
	if tmp != nil {
		// found := tmp.(*StageAWSModel)
		log.Printf("you have asked to tear down api stage %s with mode %s\n", sc.name, sc.teardown.Mode())

		_, err := sc.client.DeleteStage(context.TODO(), &apigatewayv2.DeleteStageInput{ApiId: &apiId, StageName: &sc.name})
		if err != nil {
			log.Fatalf("failed to delete stage %s: %v\n", sc.name, err)
		}
	} else {
		log.Printf("no api stage existed called %s for api %s\n", sc.name, apiId)
	}

}

var _ corebottom.Ensurable = &stageCreator{}

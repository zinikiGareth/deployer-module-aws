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

type deploymentCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *apigatewayv2.Client
}

func (sc *deploymentCreator) Loc() *errorsink.Location {
	return sc.loc
}

func (sc *deploymentCreator) ShortDescription() string {
	return "aws.gateway.Deployment[" + sc.name + "]"
}

func (sc *deploymentCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.gateway.Deployment[")
	iw.AttrsWhere(sc)
	iw.TextAttr("name", sc.name)
	iw.EndAttrs()
}

func (sc *deploymentCreator) CoinId() corebottom.CoinId {
	return sc.coin
}

func (sc *deploymentCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := sc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	sc.client = awsEnv.ApiGatewayV2Client()

	// Because there is no useful information about the deployment, there is nothing we can do
	pres.NotFound()
}

func (sc *deploymentCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
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

	model := &DeploymentModel{name: sc.name, loc: sc.loc, coin: sc.coin, api: apiStr}
	pres.Present(model)
}

func (sc *deploymentCreator) UpdateReality() {
	desired := sc.tools.Storage.GetCoin(sc.coin, corebottom.DETERMINE_DESIRED_MODE).(*DeploymentModel)
	created := &DeploymentAWSModel{}
	apiId := desired.api.String()

	input := &apigatewayv2.CreateDeploymentInput{ApiId: &apiId, StageName: &sc.name}
	out, err := sc.client.CreateDeployment(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to create api deployment %s: %v\n", sc.name, err)
	}
	id := *out.DeploymentId
	utils.ExponentialBackoff(func() bool {
		out, err := sc.client.GetDeployment(context.TODO(), &apigatewayv2.GetDeploymentInput{ApiId: &apiId, DeploymentId: &id})
		if err != nil {
			panic(err)
		}
		log.Printf("have status %s\n", out.DeploymentStatus)
		return out.DeploymentStatus == types.DeploymentStatusDeployed
	})
	created.name = sc.name
	created.deploymentId = *out.DeploymentId
	sc.tools.Storage.Bind(sc.coin, created)
}

func (sc *deploymentCreator) TearDown() {
	log.Printf("cannot tear down a deployment; use clean up tool\n")
}

var _ corebottom.Ensurable = &deploymentCreator{}

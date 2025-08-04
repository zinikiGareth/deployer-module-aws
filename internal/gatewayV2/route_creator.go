package gatewayV2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type routeCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	path     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *apigatewayv2.Client
}

func (rc *routeCreator) Loc() *errorsink.Location {
	return rc.loc
}

func (rc *routeCreator) ShortDescription() string {
	return "aws.gateway.Route[" + rc.path + "]"
}

func (rc *routeCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.gateway.Route[")
	iw.AttrsWhere(rc)
	iw.TextAttr("path", rc.path)
	iw.EndAttrs()
}

func (rc *routeCreator) CoinId() corebottom.CoinId {
	return rc.coin
}

func (rc *routeCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := rc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	rc.client = awsEnv.ApiGatewayV2Client()

	if !utils.HasProp(rc.props, "Api") {
		pres.NotFound()
		return
	}
	ae := utils.FindProp(rc.props, nil, "Api")
	apiStr, ok := rc.tools.Storage.EvalAsStringer(ae)
	if apiStr == nil {
		// if we can't resolve apiId, we won't be able to find it :-)
		pres.NotFound()
		return
	}
	if !ok {
		panic("not ok")
	}
	apiId := apiStr.String()

	var nextTok *string
	var wanted *types.Route
outer:
	for {
		curr, err := rc.client.GetRoutes(context.TODO(), &apigatewayv2.GetRoutesInput{ApiId: &apiId, NextToken: nextTok})
		if err != nil {
			log.Fatalf("could not recover route list: %v\n", err)
		}
		for _, rt := range curr.Items {
			if *rt.RouteKey == rc.path {
				wanted = &rt
				break outer
			}
		}
		if curr.NextToken == nil {
			break
		}
		nextTok = curr.NextToken
	}
	if wanted == nil {
		log.Printf("did not find route for %s with path %s\n", apiId, rc.path)
		pres.NotFound()
		return
	}

	log.Printf("found route %s\n", *wanted.RouteId)
	model := &RouteAWSModel{routeId: *wanted.RouteId}
	pres.Present(model)
}

func (rc *routeCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var api driverbottom.Expr
	var target driverbottom.Expr
	for p, v := range rc.props {
		switch p.Id() {
		case "Api":
			api = v
		case "Target":
			target = v
		default:
			rc.tools.Reporter.ReportAtf(p.Loc(), "invalid property for Api route: %s", p.Id())
		}
	}
	if api == nil {
		rc.tools.Reporter.ReportAtf(rc.loc, "no Api specified for Route")
		return
	}
	if target == nil {
		rc.tools.Reporter.ReportAtf(rc.loc, "no Target specified for Route")
		return
	}

	apiStr, ok := rc.tools.Storage.EvalAsStringer(api)
	if !ok {
		panic("not ok")
	}

	targetStr, ok := rc.tools.Storage.EvalAsStringer(target)
	if !ok {
		panic("not ok")
	}

	model := &RouteModel{path: rc.path, loc: rc.loc, coin: rc.coin, api: apiStr, target: targetStr}
	pres.Present(model)
}

func (rc *routeCreator) UpdateReality() {
	tmp := rc.tools.Storage.GetCoin(rc.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := rc.tools.Storage.GetCoin(rc.coin, corebottom.DETERMINE_DESIRED_MODE).(*RouteModel)
	created := &RouteAWSModel{}
	if tmp != nil {
		found := tmp.(*RouteAWSModel)
		created.routeId = found.routeId

		log.Printf("route already existed for %s: %s\n", rc.path, found.routeId)
		log.Printf("not handling diffs yet; just copying ...")
		rc.tools.Storage.Bind(rc.coin, created)
		return
	}

	apiId := desired.api.String()
	tgt := fmt.Sprintf("integrations/%s", desired.target.String())

	input := &apigatewayv2.CreateRouteInput{ApiId: &apiId, RouteKey: &rc.path, Target: &tgt}
	out, err := rc.client.CreateRoute(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to create api route %s: %v\n", rc.path, err)
	}
	log.Printf("created api route %s\n", *out.RouteId)
	created.routeId = *out.RouteId
	rc.tools.Storage.Bind(rc.coin, created)

}

func (rc *routeCreator) TearDown() {
	tmp := rc.tools.Storage.GetCoin(rc.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := rc.tools.Storage.GetCoin(rc.coin, corebottom.DETERMINE_DESIRED_MODE).(*RouteModel)

	apiId := desired.api.String()
	if tmp != nil {
		found := tmp.(*RouteAWSModel)
		log.Printf("you have asked to tear down api route %s with mode %s\n", rc.path, rc.teardown.Mode())

		_, err := rc.client.DeleteRoute(context.TODO(), &apigatewayv2.DeleteRouteInput{ApiId: &apiId, RouteId: &found.routeId})
		if err != nil {
			log.Fatalf("failed to delete route %s: %v\n", rc.path, err)
		}
	} else {
		log.Printf("no api route existed called %s for api %s\n", rc.path, apiId)
	}

}

var _ corebottom.Ensurable = &routeCreator{}

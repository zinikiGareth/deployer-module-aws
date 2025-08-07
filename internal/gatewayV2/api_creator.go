package gatewayV2

import (
	"context"
	"fmt"
	"log"

	hterr "github.com/aws/aws-sdk-go-v2/aws/transport/http"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type apiCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *apigatewayv2.Client
}

func (ac *apiCreator) Loc() *errorsink.Location {
	return ac.loc
}

func (ac *apiCreator) ShortDescription() string {
	return "api.gatewayV2.Api[" + ac.name + "]"
}

func (ac *apiCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("api.gatewayV2.Api %s", ac.name)
	iw.AttrsWhere(ac)
	iw.EndAttrs()
}

func (ac *apiCreator) CoinId() corebottom.CoinId {
	return ac.coin
}

func (ac *apiCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := ac.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	ac.client = awsEnv.ApiGatewayV2Client()

	var nextTok *string
	var wanted *types.Api
outer:
	for {
		curr, err := ac.client.GetApis(context.TODO(), &apigatewayv2.GetApisInput{NextToken: nextTok})
		if err != nil {
			log.Fatalf("could not recover api list: %v\n", err)
		}
		for _, api := range curr.Items {
			if *api.Name == ac.name {
				wanted = &api
				break outer
			}
		}
		if curr.NextToken == nil {
			break
		}
		nextTok = curr.NextToken
	}
	if wanted == nil {
		pres.NotFound()
		return
	}
	log.Printf("found api gateway %s\n", *wanted.ApiId)
	model := &ApiAWSModel{api: wanted, coin: ac.coin}
	pres.Present(model)
}

func (ac *apiCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var dualstack driverbottom.Expr
	var protocol driverbottom.Expr
	var rse driverbottom.Expr
	for p, v := range ac.props {
		switch p.Id() {
		case "IpAddressType":
			dualstack = v
		case "Protocol":
			protocol = v
		case "RouteSelectionExpression":
			rse = v
		default:
			ac.tools.Reporter.ReportAtf(p.Loc(), "invalid property for ApiGateway: %s", p.Id())
		}
	}
	if protocol == nil {
		ac.tools.Reporter.ReportAtf(ac.loc, "Protocol was not defined")
	}

	var at types.IpAddressType
	if dualstack != nil {
		ipat, ok := ac.tools.Storage.EvalAsStringer(dualstack)
		if !ok {
			panic("not ok")
		}

		switch ipat.String() {
		case "dualstack":
			at = types.IpAddressTypeDualstack
		case "ipv4":
			at = types.IpAddressTypeIpv4
		default:
			ac.tools.Reporter.ReportAtf(ac.loc, "invalid ip address type %s", ipat.String())
			return
		}
	}

	prot, ok := ac.tools.Storage.EvalAsStringer(protocol)
	if !ok {
		panic("not ok")
	}

	var pt types.ProtocolType
	switch prot.String() {
	case "http":
		pt = types.ProtocolTypeHttp
	case "websocket":
		pt = types.ProtocolTypeWebsocket
		if rse == nil {
			ac.tools.Reporter.ReportAtf(ac.loc, "websocket protocol requires RouteSelectionExpression")
			return
		}
	default:
		ac.tools.Reporter.ReportAtf(ac.loc, "invalid protocol type %s", prot.String())
		return
	}

	var route fmt.Stringer
	if rse != nil {
		route, ok = ac.tools.Storage.EvalAsStringer(rse)
		if !ok {
			panic("not ok")
		}
	}

	model := &ApiModel{name: ac.name, loc: ac.loc, coin: ac.coin, ipat: at, protocol: pt, rse: route}
	pres.Present(model)
}

func (ac *apiCreator) UpdateReality() {
	tmp := ac.tools.Storage.GetCoin(ac.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := ac.tools.Storage.GetCoin(ac.coin, corebottom.DETERMINE_DESIRED_MODE).(*ApiModel)
	created := &ApiAWSModel{}
	if tmp != nil {
		found := tmp.(*ApiAWSModel)
		input := &apigatewayv2.UpdateApiInput{ApiId: found.api.ApiId}
		if desired.rse != nil {
			s := desired.rse.String()
			if s != "" {
				input.RouteSelectionExpression = &s
			}
		}
		if desired.ipat != "" {
			input.IpAddressType = desired.ipat
		}
		out, err := ac.client.UpdateApi(context.TODO(), input)
		if err != nil {
			log.Fatalf("failed to update api %s: %v\n", ac.name, err)
		}
		created.api = &types.Api{Name: out.Name, ApiId: out.ApiId, ApiEndpoint: out.ApiEndpoint}
		log.Printf("updated api %s for %s\n", *found.api.ApiId, *found.api.Name)
		ac.tools.Storage.Bind(ac.coin, created)
		return
	}

	input := &apigatewayv2.CreateApiInput{Name: &ac.name, ProtocolType: desired.protocol}
	if desired.rse != nil {
		s := desired.rse.String()
		if s != "" {
			input.RouteSelectionExpression = &s
		}
	}
	if desired.ipat != "" {
		input.IpAddressType = desired.ipat
	}
	out, err := ac.client.CreateApi(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to create api %s: %v\n", ac.name, err)
	}
	log.Printf("created api %s %s\n", *out.ApiId, *out.ApiEndpoint)
	created.api = &types.Api{Name: out.Name, ApiId: out.ApiId, ApiEndpoint: out.ApiEndpoint}
	ac.tools.Storage.Bind(ac.coin, created)
}

func (ac *apiCreator) TearDown() {
	tmp := ac.tools.Storage.GetCoin(ac.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*ApiAWSModel)
		log.Printf("you have asked to tear down api %s with mode %s\n", ac.name, ac.teardown.Mode())

		_, err := ac.client.DeleteApi(context.TODO(), &apigatewayv2.DeleteApiInput{ApiId: found.api.ApiId})
		if err != nil {
			log.Fatalf("failed to delete api %s: %v\n", ac.name, err)
		}
	} else {
		log.Printf("no api existed called %s\n", ac.name)
	}
}

func thingExists(err error) bool {
	if err == nil {
		return true
	}
	e1, ok := err.(*smithy.OperationError)
	if ok {
		e2, ok := e1.Err.(*hterr.ResponseError)
		if ok {
			if e2.ResponseError.Response.StatusCode == 404 {
				switch e4 := e2.Err.(type) {
				case *types.NotFoundException:
					return false
				default:
					log.Printf("error: %T %v", e4, e4)
					panic("what error?")
				}
			}
			log.Fatalf("error: %T %v %T %v", e2.Response.Status, e2.Response.Status, e2.ResponseError.Response.StatusCode, e2.ResponseError.Response.StatusCode)
		}
		log.Fatalf("error: %T %v", e1.Err, e1.Err)
	}
	log.Fatalf("getting thing failed: %T %v", err, err)
	panic("failed")
}

var _ corebottom.Ensurable = &apiCreator{}

package gatewayV2

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type integrationCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *apigatewayv2.Client
}

func (ic *integrationCreator) Loc() *errorsink.Location {
	return ic.loc
}

func (ic *integrationCreator) ShortDescription() string {
	return "api.gatewayV2.Integration[" + ic.name + "]"
}

func (ic *integrationCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("api.gatewayV2.Integration %s", ic.name)
	iw.AttrsWhere(ic)
	iw.EndAttrs()
}

func (ic *integrationCreator) CoinId() corebottom.CoinId {
	return ic.coin
}

func (ic *integrationCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := ic.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	ic.client = awsEnv.ApiGatewayV2Client()

	if !utils.HasProp(ic.props, "Api") {
		pres.NotFound()
		return
	}

	ae := utils.FindProp(ic.props, nil, "Api")
	apiStr, ok := ic.tools.Storage.EvalAsStringer(ae)
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
	var wanted *types.Integration
	dname := fmt.Sprintf("zd[%s]", ic.name)
outer:
	for {
		curr, err := ic.client.GetIntegrations(context.TODO(), &apigatewayv2.GetIntegrationsInput{ApiId: &apiId, NextToken: nextTok})
		if err != nil {
			log.Fatalf("could not recover integration list: %v\n", err)
		}
		for _, intg := range curr.Items {
			if intg.Description != nil && *intg.Description == dname {
				wanted = &intg
				break outer
			}
		}
		if curr.NextToken == nil {
			break
		}
		nextTok = curr.NextToken
	}
	if wanted == nil {
		log.Printf("did not find integration for %s called %s\n", apiId, ic.name)
		pres.NotFound()
		return
	}

	log.Printf("found integration %s\n", *wanted.IntegrationId)
	model := &IntegrationAWSModel{integration: wanted, coin: ic.coin}
	pres.Present(model)
}

func (ic *integrationCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var api driverbottom.Expr
	var connId driverbottom.Expr
	var connType driverbottom.Expr
	var itype driverbottom.Expr
	var pfv driverbottom.Expr
	var region driverbottom.Expr
	var uri driverbottom.Expr
	for p, v := range ic.props {
		switch p.Id() {
		case "Description":
			ic.tools.Reporter.ReportAtf(ic.loc, "Description is not allowed for Integration because we use it for Name")
		case "Api":
			api = v
		case "ConnectionId":
			connId = v
		case "ConnectionType":
			connType = v
		case "PayloadFormatVersion":
			pfv = v
		case "Region":
			region = v
		case "Type":
			itype = v
		case "Uri":
			uri = v
		default:
			ic.tools.Reporter.ReportAtf(p.Loc(), "invalid property for Api Integration: %s", p.Id())
		}
	}
	if api == nil {
		ic.tools.Reporter.ReportAtf(ic.loc, "no Api specified for Integration")
		return
	}
	if region == nil {
		ic.tools.Reporter.ReportAtf(ic.loc, "no Region specified for Integration")
		return
	}
	if itype == nil {
		ic.tools.Reporter.ReportAtf(ic.loc, "no Type specified for Integration")
		return
	}
	if uri == nil {
		ic.tools.Reporter.ReportAtf(ic.loc, "no Uri specified for Integration")
		return
	}

	apiStr, ok := ic.tools.Storage.EvalAsStringer(api)
	if !ok {
		panic("not ok")
	}

	var pfvStr fmt.Stringer = nil
	if pfv != nil {
		pfvStr, ok = ic.tools.Storage.EvalAsStringer(pfv)
		if !ok {
			panic("not ok")
		}

	}

	regionStr, ok := ic.tools.Storage.EvalAsStringer(region)
	if !ok {
		panic("not ok")
	}

	typeStr, ok := ic.tools.Storage.EvalAsStringer(itype)
	if !ok {
		panic("not ok")
	}

	_, ok = ic.tools.Storage.EvalAsStringer(uri)
	if !ok {
		log.Printf("could not evaluate uri in mode 2")
	}

	var cType fmt.Stringer
	if connType != nil {
		cType, ok = ic.tools.Storage.EvalAsStringer(connType)
		if !ok {
			panic("not ok")
		}
	}
	var cId fmt.Stringer
	if connId != nil {
		cId, ok = ic.tools.Storage.EvalAsStringer(connId)
		if !ok {
			panic("not ok")
		}
	}

	model := &IntegrationModel{name: ic.name, loc: ic.loc, coin: ic.coin, api: apiStr, region: regionStr, itype: typeStr, pfv: pfvStr, uri: uri, connType: cType, connId: cId}
	pres.Present(model)
}

func (ic *integrationCreator) UpdateReality() {
	tmp := ic.tools.Storage.GetCoin(ic.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := ic.tools.Storage.GetCoin(ic.coin, corebottom.DETERMINE_DESIRED_MODE).(*IntegrationModel)
	apiId := desired.api.String()
	var itype types.IntegrationType
	switch desired.itype.String() {
	case "aws":
		fallthrough
	case "AWS":
		fallthrough
	case "lambda":
		itype = types.IntegrationTypeAws
	case "aws_proxy":
		fallthrough
	case "AWS_PROXY":
		fallthrough
	case "lambda_proxy":
		itype = types.IntegrationTypeAwsProxy
	default:
		ic.tools.Reporter.ReportAtf(ic.loc, "invalid integration type: %s\n", desired.itype.String())
	}
	uriStr, ok := ic.tools.Storage.EvalAsStringer(desired.uri)
	if !ok {
		panic("not ok")
	}
	uri := uriStr.String()

	dname := fmt.Sprintf("zd[%s]", ic.name)
	var pfv *string

	if desired.pfv != nil {
		pfv = aws.String(desired.pfv.String())
	}

	created := &IntegrationAWSModel{}
	if tmp != nil {
		found := tmp.(*IntegrationAWSModel)
		created.integration = found.integration
		log.Printf("integration already existed for %s: %s\n", ic.name, *found.integration.IntegrationId)

		input := &apigatewayv2.UpdateIntegrationInput{Description: &dname, ApiId: &apiId, IntegrationId: found.integration.IntegrationId, IntegrationType: types.IntegrationType(itype), IntegrationUri: &uri, PayloadFormatVersion: pfv}
		out, err := ic.client.UpdateIntegration(context.TODO(), input)
		if err != nil {
			log.Fatalf("failed to update api integration %s: %v\n", ic.name, err)
		}

		log.Printf("updated api integration %s\n", *out.IntegrationId)
		created.integration = &types.Integration{IntegrationId: out.IntegrationId}
		ic.tools.Storage.Bind(ic.coin, created)
		return
	}

	input := &apigatewayv2.CreateIntegrationInput{Description: &dname, ApiId: &apiId, IntegrationType: types.IntegrationType(itype), IntegrationUri: &uri, PayloadFormatVersion: pfv}
	if desired.connType != nil {
		input.ConnectionType = types.ConnectionType(desired.connType.String())
	}
	if desired.connId != nil {
		input.ConnectionId = aws.String(desired.connType.String())
	}
	out, err := ic.client.CreateIntegration(context.TODO(), input)
	if err != nil {
		log.Fatalf("failed to create api integration %s: %v\n", ic.name, err)
	}
	log.Printf("created api integration %s\n", *out.IntegrationId)
	created.integration = &types.Integration{IntegrationId: out.IntegrationId}
	ic.tools.Storage.Bind(ic.coin, created)
}

func (ic *integrationCreator) TearDown() {
	tmp := ic.tools.Storage.GetCoin(ic.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := ic.tools.Storage.GetCoin(ic.coin, corebottom.DETERMINE_DESIRED_MODE).(*IntegrationModel)

	apiId := desired.api.String()
	if tmp != nil {
		found := tmp.(*IntegrationAWSModel)
		log.Printf("you have asked to tear down api integration %s with mode %s\n", ic.name, ic.teardown.Mode())

		_, err := ic.client.DeleteIntegration(context.TODO(), &apigatewayv2.DeleteIntegrationInput{ApiId: &apiId, IntegrationId: found.integration.IntegrationId})
		if err != nil {
			log.Fatalf("failed to delete integration %s: %v\n", ic.name, err)
		}
	} else {
		log.Printf("no api integration existed called %s for api %s\n", ic.name, apiId)
	}
}

var _ corebottom.Ensurable = &integrationCreator{}

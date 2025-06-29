package cfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type RHPCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *cloudfront.Client
}

func (rhpc *RHPCreator) Loc() *errorsink.Location {
	return rhpc.loc
}

func (rhpc *RHPCreator) ShortDescription() string {
	return "aws.CloudFront.RHP[" + rhpc.name + "]"
}

func (rhpc *RHPCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.CloudFront.RHP[")
	iw.AttrsWhere(rhpc)
	iw.TextAttr("named", rhpc.name)
	if rhpc.teardown != nil {
		iw.TextAttr("teardown", rhpc.teardown.Mode())
	}
	iw.EndAttrs()
}

func (rhpc *RHPCreator) CoinId() corebottom.CoinId {
	return rhpc.coin
}

func (rhpc *RHPCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := rhpc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	rhpc.client = awsEnv.CFClient()

	zeb, err := rhpc.client.ListResponseHeadersPolicies(context.TODO(), &cloudfront.ListResponseHeadersPoliciesInput{})
	if err != nil {
		log.Fatalf("could not list RHPs")
	}
	model := &rhpModel{loc: rhpc.loc, name: rhpc.name, coin: rhpc.coin}
	found := false
	for _, p := range zeb.ResponseHeadersPolicyList.Items {
		if p.ResponseHeadersPolicy.Id != nil {
			rhc, err := rhpc.client.GetResponseHeadersPolicyConfig(context.TODO(), &cloudfront.GetResponseHeadersPolicyConfigInput{Id: p.ResponseHeadersPolicy.Id})
			if err != nil {
				log.Fatalf("could not recover RHP %s", *p.ResponseHeadersPolicy.Id)
			}
			if rhc.ResponseHeadersPolicyConfig.Name != nil && *rhc.ResponseHeadersPolicyConfig.Name == rhpc.name {
				model.rpId = *p.ResponseHeadersPolicy.Id
				log.Printf("found rhpc %s\n", model.rpId)
				found = true
			}
		}
	}

	if found {
		pres.Present(model)
	} else {
		pres.NotFound()
	}
}

func (rhpc *RHPCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var header driverbottom.Expr
	var value driverbottom.Expr
	for p, v := range rhpc.props {
		switch p.Id() {
		case "Header":
			header = v
		case "Value":
			value = v
		default:
			rhpc.tools.Reporter.ReportAtf(rhpc.loc, "invalid property for ResponseHeaderPolicy: %s", p.Id())
		}
	}

	model := &rhpModel{loc: rhpc.loc, name: rhpc.name, coin: rhpc.coin, header: header, value: value}
	pres.Present(model)
}

func (rhpc *RHPCreator) UpdateReality() {
	tmp := rhpc.tools.Storage.GetCoin(rhpc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*rhpModel)
		log.Printf("RHP %s already existed for %s\n", found.rpId, found.name)
		return
	}

	desired := rhpc.tools.Storage.GetCoin(rhpc.coin, corebottom.DETERMINE_DESIRED_MODE).(*rhpModel)

	created := &rhpModel{loc: desired.loc, name: desired.name, coin: rhpc.coin}

	ht, ok1 := rhpc.tools.Storage.EvalAsStringer(desired.header)
	ov := true
	vt, ok2 := rhpc.tools.Storage.EvalAsStringer(desired.value)
	if !ok1 || !ok2 {
		panic("not ok")
	}
	h := ht.String()
	v := vt.String()
	rhs := []types.ResponseHeadersPolicyCustomHeader{{Header: &h, Override: &ov, Value: &v}}
	rhslen := int32(len(rhs))
	ch := types.ResponseHeadersPolicyCustomHeadersConfig{Items: rhs, Quantity: &rhslen}
	rhp := types.ResponseHeadersPolicyConfig{Name: &rhpc.name, CustomHeadersConfig: &ch}
	crhp, err := rhpc.client.CreateResponseHeadersPolicy(context.TODO(), &cloudfront.CreateResponseHeadersPolicyInput{ResponseHeadersPolicyConfig: &rhp})
	if err != nil {
		log.Fatalf("failed to create CRHP %s: %v\n", rhpc.name, err)
	}
	created.rpId = *crhp.ResponseHeadersPolicy.Id
	log.Printf("created RHP for %s: %s\n", created.name, created.rpId)

	rhpc.tools.Storage.Bind(rhpc.coin, created)
}

func (rhpc *RHPCreator) TearDown() {
	tmp := rhpc.tools.Storage.GetCoin(rhpc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*rhpModel)
		log.Printf("you have asked to tear down RHP %s (id: %s) with mode %s\n", found.name, found.rpId, rhpc.teardown.Mode())
		x, err := rhpc.client.GetResponseHeadersPolicy(context.TODO(), &cloudfront.GetResponseHeadersPolicyInput{Id: &found.rpId})
		if err != nil {
			log.Fatalf("could not get RHP %s: %v", found.rpId, err)
		}
		_, err = rhpc.client.DeleteResponseHeadersPolicy(context.TODO(), &cloudfront.DeleteResponseHeadersPolicyInput{Id: &found.rpId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete RHP %s: %v", found.rpId, err)
		}
		log.Printf("deleted RHP %s\n", found.rpId)
	} else {
		log.Printf("no RHP existed for %s\n", rhpc.name)
	}
}

var _ corebottom.Ensurable = &RHPCreator{}

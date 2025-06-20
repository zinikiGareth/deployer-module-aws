package cfront

import (
	"context"
	"fmt"
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
	header   driverbottom.Expr
	value    driverbottom.Expr
	teardown corebottom.TearDown

	client        *cloudfront.Client
	rpId          string
	alreadyExists bool
}

func (cfdc *RHPCreator) Loc() *errorsink.Location {
	return cfdc.loc
}

func (cfdc *RHPCreator) ShortDescription() string {
	return "aws.CloudFront.RHP[" + cfdc.name + "]"
}

func (cfdc *RHPCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.CloudFront.RHP[")
	iw.AttrsWhere(cfdc)
	iw.TextAttr("named", cfdc.name)
	iw.NestedAttr("header", cfdc.header)
	iw.NestedAttr("value", cfdc.value)
	if cfdc.teardown != nil {
		iw.TextAttr("teardown", cfdc.teardown.Mode())
	}
	iw.EndAttrs()
}

func (cfdc *RHPCreator) BuildModel(pres driverbottom.ValuePresenter) {
	eq := cfdc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	cfdc.client = awsEnv.CFClient()

	zeb, err := cfdc.client.ListResponseHeadersPolicies(context.TODO(), &cloudfront.ListResponseHeadersPoliciesInput{})
	if err != nil {
		log.Fatalf("could not list RHPs")
	}
	for _, p := range zeb.ResponseHeadersPolicyList.Items {
		if p.ResponseHeadersPolicy.Id != nil {
			rhc, err := cfdc.client.GetResponseHeadersPolicyConfig(context.TODO(), &cloudfront.GetResponseHeadersPolicyConfigInput{Id: p.ResponseHeadersPolicy.Id})
			if err != nil {
				log.Fatalf("could not recover RHP %s", *p.ResponseHeadersPolicy.Id)
			}
			if rhc.ResponseHeadersPolicyConfig.Name != nil && *rhc.ResponseHeadersPolicyConfig.Name == cfdc.name {
				cfdc.rpId = *p.ResponseHeadersPolicy.Id
				cfdc.alreadyExists = true
				log.Printf("found rhpc %s\n", cfdc.rpId)
			}
		}
	}

	pres.Present(cfdc)
}

func (cfdc *RHPCreator) UpdateReality() {
	if cfdc.alreadyExists {
		log.Printf("RHP %s already existed for %s\n", cfdc.rpId, cfdc.name)
		return
	}
	h := cfdc.tools.Storage.EvalAsStringer(cfdc.header).String()
	ov := true
	v := cfdc.tools.Storage.EvalAsStringer(cfdc.value).String()
	rhs := []types.ResponseHeadersPolicyCustomHeader{{Header: &h, Override: &ov, Value: &v}}
	rhslen := int32(len(rhs))
	ch := types.ResponseHeadersPolicyCustomHeadersConfig{Items: rhs, Quantity: &rhslen}
	rhp := types.ResponseHeadersPolicyConfig{Name: &cfdc.name, CustomHeadersConfig: &ch}
	crhp, err := cfdc.client.CreateResponseHeadersPolicy(context.TODO(), &cloudfront.CreateResponseHeadersPolicyInput{ResponseHeadersPolicyConfig: &rhp})
	if err != nil {
		log.Fatalf("failed to create CRHP %s: %v\n", cfdc.name, err)
	}
	cfdc.rpId = *crhp.ResponseHeadersPolicy.Id
	log.Printf("created RHP for %s: %s\n", cfdc.name, cfdc.rpId)
}

func (cfdc *RHPCreator) TearDown() {
	if cfdc.alreadyExists {
		log.Printf("you have asked to tear down RHP %s (id: %s) with mode %s\n", cfdc.name, cfdc.rpId, cfdc.teardown.Mode())
		x, err := cfdc.client.GetResponseHeadersPolicy(context.TODO(), &cloudfront.GetResponseHeadersPolicyInput{Id: &cfdc.rpId})
		if err != nil {
			log.Fatalf("could not get RHP %s: %v", cfdc.rpId, err)
		}
		_, err = cfdc.client.DeleteResponseHeadersPolicy(context.TODO(), &cloudfront.DeleteResponseHeadersPolicyInput{Id: &cfdc.rpId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete RHP %s: %v", cfdc.rpId, err)
		}
		log.Printf("deleted RHP %s\n", cfdc.rpId)
	} else {
		log.Printf("no RHP existed for %s\n", cfdc.name)
	}
}

func (cfdc *RHPCreator) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "id":
		return &RHPIdMethod{}
	}
	return nil
}

type RHPIdMethod struct {
}

func (a *RHPIdMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	cfdc, ok := e.(*RHPCreator)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a RHP, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if cfdc.alreadyExists {
		return cfdc.rpId
	} else {
		return &deferReadingRHPId{cfdc: cfdc}
	}
}

type deferReadingRHPId struct {
	cfdc *RHPCreator
}

func (d *deferReadingRHPId) String() string {
	if d.cfdc.rpId == "" {
		panic("id is still not set")
	}
	return d.cfdc.rpId
}

var _ driverbottom.HasMethods = &RHPCreator{}

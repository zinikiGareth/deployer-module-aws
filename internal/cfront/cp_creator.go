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
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type CachePolicyCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	minttl   driverbottom.Expr
	teardown corebottom.TearDown

	client        *cloudfront.Client
	CachePolicyId string
	alreadyExists bool
}

func (cfdc *CachePolicyCreator) Loc() *errorsink.Location {
	return cfdc.loc
}

func (cfdc *CachePolicyCreator) ShortDescription() string {
	return "aws.CloudFront.CachePolicy[" + cfdc.name + "]"
}

func (cfdc *CachePolicyCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.CloudFront.CachePolicy[")
	iw.AttrsWhere(cfdc)
	iw.TextAttr("named", cfdc.name)
	iw.NestedAttr("minttl", cfdc.minttl)
	if cfdc.teardown != nil {
		iw.TextAttr("teardown", cfdc.teardown.Mode())
	}
	iw.EndAttrs()
}

func (cfdc *CachePolicyCreator) BuildModel(pres driverbottom.ValuePresenter) {
	eq := cfdc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	cfdc.client = awsEnv.CFClient()

	bert, err := cfdc.client.ListCachePolicies(context.TODO(), &cloudfront.ListCachePoliciesInput{})
	if err != nil {
		log.Fatalf("could not list CPs")
	}
	for _, p := range bert.CachePolicyList.Items {
		if p.CachePolicy.Id != nil && p.CachePolicy.CachePolicyConfig.Name != nil && *p.CachePolicy.CachePolicyConfig.Name == cfdc.name {
			cfdc.CachePolicyId = *p.CachePolicy.Id
			cfdc.alreadyExists = true
			log.Printf("found CachePolicy for %s with id %s\n", cfdc.name, cfdc.CachePolicyId)
		}
	}

	pres.Present(cfdc)
}

func (cfdc *CachePolicyCreator) UpdateReality() {
	if cfdc.alreadyExists {
		log.Printf("CachePolicy %s already existed for %s\n", cfdc.CachePolicyId, cfdc.name)
		return
	}
	mt := cfdc.tools.Storage.Eval(cfdc.minttl)
	minttlVal, ok := mt.(float64)
	if !ok {
		log.Fatalf("not float64 but %T", mt)
	}
	var minttl int64 = int64(minttlVal)
	cpc := types.CachePolicyConfig{Name: &cfdc.name, MinTTL: &minttl}
	oac, err := cfdc.client.CreateCachePolicy(context.TODO(), &cloudfront.CreateCachePolicyInput{CachePolicyConfig: &cpc})
	if err != nil {
		log.Fatalf("failed to create CachePolicy for %s: %v\n", cfdc.name, err)
	}
	cfdc.CachePolicyId = *oac.CachePolicy.Id
	log.Printf("created CachePolicy for %s: %s\n", cfdc.name, cfdc.CachePolicyId)
}

func (cfdc *CachePolicyCreator) TearDown() {
	if cfdc.alreadyExists {
		log.Printf("you have asked to tear down CachePolicy %s (id: %s) with mode %s\n", cfdc.name, cfdc.CachePolicyId, cfdc.teardown.Mode())
		x, err := cfdc.client.GetCachePolicy(context.TODO(), &cloudfront.GetCachePolicyInput{Id: &cfdc.CachePolicyId})
		if err != nil {
			log.Fatalf("could not get CP %s: %v", cfdc.CachePolicyId, err)
		}
		_, err = cfdc.client.DeleteCachePolicy(context.TODO(), &cloudfront.DeleteCachePolicyInput{Id: &cfdc.CachePolicyId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete CP %s: %v", cfdc.CachePolicyId, err)
		}
		log.Printf("deleted CachePolicy %s\n", cfdc.CachePolicyId)
	} else {
		log.Printf("no CachePolicy existed for %s\n", cfdc.name)
	}
}

func (cfdc *CachePolicyCreator) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "id":
		return &CachePolicyIdMethod{}
	}
	return nil
}

type CachePolicyIdMethod struct {
}

func (a *CachePolicyIdMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	cfdc, ok := e.(*CachePolicyCreator)
	if !ok {
		panic(fmt.Sprintf("id can only be called on a CachePolicy, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if cfdc.alreadyExists {
		return cfdc.CachePolicyId
	} else {
		return utils.DeferString(func() string {
			if cfdc.CachePolicyId == "" {
				panic("id is still not set")
			}
			return cfdc.CachePolicyId
		})
	}
}

var _ driverbottom.HasMethods = &CachePolicyCreator{}

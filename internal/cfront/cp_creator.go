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

type CachePolicyCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	teardown corebottom.TearDown
	props    map[driverbottom.Identifier]driverbottom.Expr

	client *cloudfront.Client
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
	if cfdc.teardown != nil {
		iw.TextAttr("teardown", cfdc.teardown.Mode())
	}
	iw.EndAttrs()
}

func (acmc *CachePolicyCreator) CoinId() corebottom.CoinId {
	return acmc.coin
}

func (cfdc *CachePolicyCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := cfdc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	cfdc.client = awsEnv.CFClient()

	var model *cachePolicyModel
	bert, err := cfdc.client.ListCachePolicies(context.TODO(), &cloudfront.ListCachePoliciesInput{})
	if err != nil {
		log.Fatalf("could not list CPs")
	}
	for _, p := range bert.CachePolicyList.Items {
		if p.CachePolicy.Id != nil && p.CachePolicy.CachePolicyConfig.Name != nil && *p.CachePolicy.CachePolicyConfig.Name == cfdc.name {
			model = NewCachePolicyModel(cfdc.coin, cfdc.loc, cfdc.name, "found")
			model.CachePolicyId = *p.CachePolicy.Id
			log.Printf("found CachePolicy for %s with id %s\n", model.name, model.CachePolicyId)
		}
	}
	if model != nil {
		pres.Present(model)
	} else {
		pres.NotFound()
	}
}

func (cfdc *CachePolicyCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var minttl driverbottom.Expr
	for p, v := range cfdc.props {
		switch p.Id() {
		case "MinTTL":
			minttl = v
		default:
			cfdc.tools.Reporter.ReportAtf(cfdc.loc, "invalid property for OriginAccessControl: %s", p.Id())
		}
	}

	model := NewCachePolicyModel(cfdc.coin, cfdc.loc, cfdc.name, "desired")
	model.minttl = minttl

	// cfdc.tools.Storage.Bind(cfdc.coin, model)
	pres.Present(model)
}

func (cfdc *CachePolicyCreator) UpdateReality() {
	tmp := cfdc.tools.Storage.GetCoin(cfdc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*cachePolicyModel)
		log.Printf("CachePolicy %s already existed for %s\n", found.CachePolicyId, found.name)
		return
	}

	desired := cfdc.tools.Storage.GetCoin(cfdc.coin, corebottom.DETERMINE_DESIRED_MODE).(*cachePolicyModel)

	created := NewCachePolicyModel(cfdc.coin, desired.loc, desired.name, "created")

	created.minttl = desired.minttl

	mt := cfdc.tools.Storage.EvalAsNumber(created.minttl)
	minttl := int64(mt.F64())
	cpc := types.CachePolicyConfig{Name: &cfdc.name, MinTTL: &minttl}
	oac, err := cfdc.client.CreateCachePolicy(context.TODO(), &cloudfront.CreateCachePolicyInput{CachePolicyConfig: &cpc})
	if err != nil {
		log.Fatalf("failed to create CachePolicy for %s: %v\n", cfdc.name, err)
	}
	created.CachePolicyId = *oac.CachePolicy.Id
	log.Printf("created CachePolicy for %s: %s\n", cfdc.name, created.CachePolicyId)

	cfdc.tools.Storage.Bind(cfdc.coin, created)
}

func (cfdc *CachePolicyCreator) TearDown() {
	tmp := cfdc.tools.Storage.GetCoin(cfdc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*cachePolicyModel)
		log.Printf("you have asked to tear down CachePolicy %s (id: %s) with mode %s\n", cfdc.name, found.CachePolicyId, cfdc.teardown.Mode())
		x, err := cfdc.client.GetCachePolicy(context.TODO(), &cloudfront.GetCachePolicyInput{Id: &found.CachePolicyId})
		if err != nil {
			log.Fatalf("could not get CP %s: %v", found.CachePolicyId, err)
		}
		_, err = cfdc.client.DeleteCachePolicy(context.TODO(), &cloudfront.DeleteCachePolicyInput{Id: &found.CachePolicyId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete CP %s: %v", found.CachePolicyId, err)
		}
		log.Printf("deleted CachePolicy %s\n", found.CachePolicyId)
	} else {
		log.Printf("no CachePolicy existed for %s\n", cfdc.name)
	}
}

var _ corebottom.Ensurable = &CachePolicyCreator{}

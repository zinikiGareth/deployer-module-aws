package cfront

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type CachePolicyCreator struct {
	tools *pluggable.Tools

	loc          *errorsink.Location
	name         string
	acType       pluggable.Expr
	signBehavior pluggable.Expr
	signProt     pluggable.Expr
	teardown     pluggable.TearDown

	client        *cloudfront.Client
	CachePolicyId         string
	alreadyExists bool
}

func (cfdc *CachePolicyCreator) Loc() *errorsink.Location {
	return cfdc.loc
}

func (cfdc *CachePolicyCreator) ShortDescription() string {
	return "aws.CloudFront.CachePolicy[" + cfdc.name + "]"
}

func (cfdc *CachePolicyCreator) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.CloudFront.CachePolicy[")
	iw.AttrsWhere(cfdc)
	iw.TextAttr("named", cfdc.name)
	iw.NestedAttr("acType", cfdc.acType)
	iw.NestedAttr("signingBehavior", cfdc.signBehavior)
	iw.NestedAttr("signingProtocol", cfdc.signProt)
	if cfdc.teardown != nil {
		iw.TextAttr("teardown", cfdc.teardown.Mode())
	}
	iw.EndAttrs()
}

func (cfdc *CachePolicyCreator) BuildModel(pres pluggable.ValuePresenter) {
	eq := cfdc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	cfdc.client = awsEnv.CFClient()

	fred, err := cfdc.client.ListOriginAccessControls(context.TODO(), &cloudfront.ListOriginAccessControlsInput{})
	if err != nil {
		log.Fatalf("could not list CachePolicys")
	}
	for _, p := range fred.OriginAccessControlList.Items {
		if p.Id != nil && p.Name != nil && *p.Name == cfdc.name {
			cfdc.CachePolicyId = *p.Id
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
	ty := types.OriginAccessControlOriginTypes(cfdc.tools.Storage.EvalAsString(cfdc.acType))
	sb := types.OriginAccessControlSigningBehaviors(cfdc.tools.Storage.EvalAsString(cfdc.signBehavior))
	sp := types.OriginAccessControlSigningProtocols(cfdc.tools.Storage.EvalAsString(cfdc.signProt))
	CachePolicycfg := types.OriginAccessControlConfig{Name: &cfdc.name, OriginAccessControlOriginType: ty, SigningBehavior: sb, SigningProtocol: sp}
	CachePolicy, err := cfdc.client.CreateOriginAccessControl(context.TODO(), &cloudfront.CreateOriginAccessControlInput{OriginAccessControlConfig: &CachePolicycfg})
	if err != nil {
		log.Fatalf("failed to create CachePolicy for %s: %v\n", cfdc.name, err)
	}
	cfdc.CachePolicyId = *CachePolicy.OriginAccessControl.Id
	log.Printf("created CachePolicy for %s: %s\n", cfdc.name, cfdc.CachePolicyId)
}

func (cfdc *CachePolicyCreator) TearDown() {
	if cfdc.alreadyExists {
		log.Printf("you have asked to tear down CachePolicy %s (id: %s) with mode %s\n", cfdc.name, cfdc.CachePolicyId, cfdc.teardown.Mode())
		x, err := cfdc.client.GetOriginAccessControl(context.TODO(), &cloudfront.GetOriginAccessControlInput{Id: &cfdc.CachePolicyId})
		if err != nil {
			log.Fatalf("could not get CachePolicy %s: %v", cfdc.CachePolicyId, err)
		}
		_, err = cfdc.client.DeleteOriginAccessControl(context.TODO(), &cloudfront.DeleteOriginAccessControlInput{Id: &cfdc.CachePolicyId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete CachePolicy %s: %v", cfdc.CachePolicyId, err)
		}
		log.Printf("deleted CachePolicy %s\n", cfdc.CachePolicyId)
	} else {
		log.Printf("no CachePolicy existed for %s\n", cfdc.name)
	}
}

func (cfdc *CachePolicyCreator) ObtainMethod(name string) pluggable.Method {
	switch name {
	case "id":
		return &CachePolicyIdMethod{}
	}
	return nil
}

type CachePolicyIdMethod struct {
}

func (a *CachePolicyIdMethod) Invoke(s pluggable.RuntimeStorage, on pluggable.Expr, args []pluggable.Expr) any {
	e := on.Eval(s)
	cfdc, ok := e.(*CachePolicyCreator)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a CachePolicy, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if cfdc.alreadyExists {
		return cfdc.CachePolicyId
	} else {
		return &deferReadingCachePolicyId{cfdc: cfdc}
	}
}

type deferReadingCachePolicyId struct {
	cfdc *CachePolicyCreator
}

func (d *deferReadingCachePolicyId) String() string {
	if d.cfdc.CachePolicyId == "" {
		panic("id is still not set")
	}
	return d.cfdc.CachePolicyId
}

var _ pluggable.HasMethods = &CachePolicyCreator{}

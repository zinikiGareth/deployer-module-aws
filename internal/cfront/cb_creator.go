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

type CacheBehaviorCreator struct {
	tools *pluggable.Tools

	loc          *errorsink.Location
	name         string
	acType       pluggable.Expr
	signBehavior pluggable.Expr
	signProt     pluggable.Expr
	teardown     pluggable.TearDown

	client        *cloudfront.Client
	CacheBehaviorId         string
	alreadyExists bool
}

func (cfdc *CacheBehaviorCreator) Loc() *errorsink.Location {
	return cfdc.loc
}

func (cfdc *CacheBehaviorCreator) ShortDescription() string {
	return "aws.CloudFront.CacheBehavior[" + cfdc.name + "]"
}

func (cfdc *CacheBehaviorCreator) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.CloudFront.CacheBehavior[")
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

func (cfdc *CacheBehaviorCreator) BuildModel(pres pluggable.ValuePresenter) {
	eq := cfdc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	cfdc.client = awsEnv.CFClient()

	fred, err := cfdc.client.ListOriginAccessControls(context.TODO(), &cloudfront.ListOriginAccessControlsInput{})
	if err != nil {
		log.Fatalf("could not list CacheBehaviors")
	}
	for _, p := range fred.OriginAccessControlList.Items {
		if p.Id != nil && p.Name != nil && *p.Name == cfdc.name {
			cfdc.CacheBehaviorId = *p.Id
			log.Printf("found CacheBehavior for %s with id %s\n", cfdc.name, cfdc.CacheBehaviorId)
		}
	}

	pres.Present(cfdc)
}

func (cfdc *CacheBehaviorCreator) UpdateReality() {
	if cfdc.alreadyExists {
		log.Printf("CacheBehavior %s already existed for %s\n", cfdc.CacheBehaviorId, cfdc.name)
		return
	}
	ty := types.OriginAccessControlOriginTypes(cfdc.tools.Storage.EvalAsString(cfdc.acType))
	sb := types.OriginAccessControlSigningBehaviors(cfdc.tools.Storage.EvalAsString(cfdc.signBehavior))
	sp := types.OriginAccessControlSigningProtocols(cfdc.tools.Storage.EvalAsString(cfdc.signProt))
	CacheBehaviorcfg := types.OriginAccessControlConfig{Name: &cfdc.name, OriginAccessControlOriginType: ty, SigningBehavior: sb, SigningProtocol: sp}
	CacheBehavior, err := cfdc.client.CreateOriginAccessControl(context.TODO(), &cloudfront.CreateOriginAccessControlInput{OriginAccessControlConfig: &CacheBehaviorcfg})
	if err != nil {
		log.Fatalf("failed to create CacheBehavior for %s: %v\n", cfdc.name, err)
	}
	cfdc.CacheBehaviorId = *CacheBehavior.OriginAccessControl.Id
	log.Printf("created CacheBehavior for %s: %s\n", cfdc.name, cfdc.CacheBehaviorId)
}

func (cfdc *CacheBehaviorCreator) TearDown() {
	if cfdc.alreadyExists {
		log.Printf("you have asked to tear down CacheBehavior %s (id: %s) with mode %s\n", cfdc.name, cfdc.CacheBehaviorId, cfdc.teardown.Mode())
		x, err := cfdc.client.GetOriginAccessControl(context.TODO(), &cloudfront.GetOriginAccessControlInput{Id: &cfdc.CacheBehaviorId})
		if err != nil {
			log.Fatalf("could not get CacheBehavior %s: %v", cfdc.CacheBehaviorId, err)
		}
		_, err = cfdc.client.DeleteOriginAccessControl(context.TODO(), &cloudfront.DeleteOriginAccessControlInput{Id: &cfdc.CacheBehaviorId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete CacheBehavior %s: %v", cfdc.CacheBehaviorId, err)
		}
		log.Printf("deleted CacheBehavior %s\n", cfdc.CacheBehaviorId)
	} else {
		log.Printf("no CacheBehavior existed for %s\n", cfdc.name)
	}
}

func (cfdc *CacheBehaviorCreator) ObtainMethod(name string) pluggable.Method {
	switch name {
	case "id":
		return &CacheBehaviorIdMethod{}
	}
	return nil
}

type CacheBehaviorIdMethod struct {
}

func (a *CacheBehaviorIdMethod) Invoke(s pluggable.RuntimeStorage, on pluggable.Expr, args []pluggable.Expr) any {
	e := on.Eval(s)
	cfdc, ok := e.(*CacheBehaviorCreator)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a CacheBehavior, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if cfdc.alreadyExists {
		return cfdc.CacheBehaviorId
	} else {
		return &deferReadingCacheBehaviorId{cfdc: cfdc}
	}
}

type deferReadingCacheBehaviorId struct {
	cfdc *CacheBehaviorCreator
}

func (d *deferReadingCacheBehaviorId) String() string {
	if d.cfdc.CacheBehaviorId == "" {
		panic("id is still not set")
	}
	return d.cfdc.CacheBehaviorId
}

var _ pluggable.HasMethods = &CacheBehaviorCreator{}

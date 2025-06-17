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

type OACCreator struct {
	tools *pluggable.Tools

	loc          *errorsink.Location
	name         string
	acType       pluggable.Expr
	signBehavior pluggable.Expr
	signProt     pluggable.Expr
	teardown     pluggable.TearDown

	client        *cloudfront.Client
	oacId         string
	alreadyExists bool
}

func (cfdc *OACCreator) Loc() *errorsink.Location {
	return cfdc.loc
}

func (cfdc *OACCreator) ShortDescription() string {
	return "aws.CloudFront.OAC[" + cfdc.name + "]"
}

func (cfdc *OACCreator) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.CloudFront.OAC[")
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

func (cfdc *OACCreator) BuildModel(pres pluggable.ValuePresenter) {
	eq := cfdc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	cfdc.client = awsEnv.CFClient()

	fred, err := cfdc.client.ListOriginAccessControls(context.TODO(), &cloudfront.ListOriginAccessControlsInput{})
	if err != nil {
		log.Fatalf("could not list OACs")
	}
	for _, p := range fred.OriginAccessControlList.Items {
		if p.Id != nil && p.Name != nil && *p.Name == cfdc.name {
			cfdc.oacId = *p.Id
			log.Printf("found OAC for %s with id %s\n", cfdc.name, cfdc.oacId)
		}
	}

	pres.Present(cfdc)
}

func (cfdc *OACCreator) UpdateReality() {
	if cfdc.alreadyExists {
		log.Printf("OAC %s already existed for %s\n", cfdc.oacId, cfdc.name)
		return
	}
	ty := types.OriginAccessControlOriginTypes(cfdc.tools.Storage.EvalAsString(cfdc.acType))
	sb := types.OriginAccessControlSigningBehaviors(cfdc.tools.Storage.EvalAsString(cfdc.signBehavior))
	sp := types.OriginAccessControlSigningProtocols(cfdc.tools.Storage.EvalAsString(cfdc.signProt))
	oaccfg := types.OriginAccessControlConfig{Name: &cfdc.name, OriginAccessControlOriginType: ty, SigningBehavior: sb, SigningProtocol: sp}
	oac, err := cfdc.client.CreateOriginAccessControl(context.TODO(), &cloudfront.CreateOriginAccessControlInput{OriginAccessControlConfig: &oaccfg})
	if err != nil {
		log.Fatalf("failed to create OAC for %s: %v\n", cfdc.name, err)
	}
	cfdc.oacId = *oac.OriginAccessControl.Id
	log.Printf("created OAC for %s: %s\n", cfdc.name, cfdc.oacId)
}

func (cfdc *OACCreator) TearDown() {
	if cfdc.alreadyExists {
		log.Printf("you have asked to tear down OAC %s (id: %s) with mode %s\n", cfdc.name, cfdc.oacId, cfdc.teardown.Mode())
		x, err := cfdc.client.GetOriginAccessControl(context.TODO(), &cloudfront.GetOriginAccessControlInput{Id: &cfdc.oacId})
		if err != nil {
			log.Fatalf("could not get OAC %s: %v", cfdc.oacId, err)
		}
		_, err = cfdc.client.DeleteOriginAccessControl(context.TODO(), &cloudfront.DeleteOriginAccessControlInput{Id: &cfdc.oacId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete OAC %s: %v", cfdc.oacId, err)
		}
		log.Printf("deleted OAC %s\n", cfdc.oacId)
	} else {
		log.Printf("no OAC existed for %s\n", cfdc.name)
	}
}

func (cfdc *OACCreator) ObtainMethod(name string) pluggable.Method {
	switch name {
	case "id":
		return &oacIdMethod{}
	}
	return nil
}

type oacIdMethod struct {
}

func (a *oacIdMethod) Invoke(s pluggable.RuntimeStorage, on pluggable.Expr, args []pluggable.Expr) any {
	e := on.Eval(s)
	cfdc, ok := e.(*OACCreator)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a OAC, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if cfdc.alreadyExists {
		return cfdc.oacId
	} else {
		return &deferReadingOACId{cfdc: cfdc}
	}
}

type deferReadingOACId struct {
	cfdc *OACCreator
}

func (d *deferReadingOACId) String() string {
	if d.cfdc.oacId == "" {
		panic("id is still not set")
	}
	return d.cfdc.oacId
}

var _ pluggable.HasMethods = &OACCreator{}

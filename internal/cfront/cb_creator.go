package cfront

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type CacheBehaviorCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	cpId     driverbottom.Expr
	pp       driverbottom.Expr
	rhp      driverbottom.Expr
	toid     driverbottom.Expr
	teardown corebottom.TearDown

	client *cloudfront.Client
}

func (cfdc *CacheBehaviorCreator) Loc() *errorsink.Location {
	return cfdc.loc
}

func (cfdc *CacheBehaviorCreator) ShortDescription() string {
	return "aws.CloudFront.CacheBehavior[" + cfdc.name + "]"
}

func (cfdc *CacheBehaviorCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.CloudFront.CacheBehavior[")
	iw.AttrsWhere(cfdc)
	iw.TextAttr("named", cfdc.name)
	iw.NestedAttr("pp", cfdc.pp)
	iw.NestedAttr("rhp", cfdc.rhp)
	if cfdc.teardown != nil {
		iw.TextAttr("teardown", cfdc.teardown.Mode())
	}
	iw.EndAttrs()
}

func (cfdc *CacheBehaviorCreator) DetermineDesiredState(pres driverbottom.ValuePresenter) {
	eq := cfdc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	cfdc.client = awsEnv.CFClient()

	pp := cfdc.tools.Storage.Eval(cfdc.pp)
	rhp := cfdc.tools.Storage.Eval(cfdc.rhp)
	targetOriginId := cfdc.tools.Storage.Eval(cfdc.toid)

	// this is going to need to handle "deferred"
	cpId, ok := cfdc.tools.Storage.EvalAsStringer(cfdc.cpId)
	if !ok {
		panic("not a string")
	}

	pres.Present(cbModel{pp: pp, rhp: rhp, targetOriginId: targetOriginId, cpId: cpId})
}

func (cfdc *CacheBehaviorCreator) UpdateReality() {
}

func (cfdc *CacheBehaviorCreator) TearDown() {
}

type cbModel struct {
	pp             any
	rhp            any
	targetOriginId any
	cpId           fmt.Stringer
}

func (d *cbModel) Complete() types.CacheBehavior {
	toi := utils.AsString(d.targetOriginId)
	pp := utils.AsString(d.pp)
	rhp := utils.AsString(d.rhp)
	cpId := d.cpId.String()
	return types.CacheBehavior{TargetOriginId: &toi, PathPattern: &pp, ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps, CachePolicyId: &cpId, ResponseHeadersPolicyId: &rhp}
}

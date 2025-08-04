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

type OACCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	teardown corebottom.TearDown

	client *cloudfront.Client
	props  map[driverbottom.Identifier]driverbottom.Expr
}

func (oacc *OACCreator) Loc() *errorsink.Location {
	return oacc.loc
}

func (oacc *OACCreator) ShortDescription() string {
	return "aws.CloudFront.OAC[" + oacc.name + "]"
}

func (oacc *OACCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.CloudFront.OAC[")
	iw.AttrsWhere(oacc)
	iw.TextAttr("named", oacc.name)
	iw.TextAttr("coin", oacc.coin.VarName().Id())
	if oacc.teardown != nil {
		iw.TextAttr("teardown", oacc.teardown.Mode())
	}
	iw.EndAttrs()
}

func (oacc *OACCreator) CoinId() corebottom.CoinId {
	return oacc.coin
}

func (oacc *OACCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := oacc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	oacc.client = awsEnv.CFClient()

	model := &oacModel{loc: oacc.loc, name: oacc.name, coin: oacc.coin}
	found := false
	fred, err := oacc.client.ListOriginAccessControls(context.TODO(), &cloudfront.ListOriginAccessControlsInput{})
	if err != nil {
		log.Fatalf("could not list OACs")
	}
	for _, p := range fred.OriginAccessControlList.Items {
		if p.Id != nil && p.Name != nil && *p.Name == oacc.name {
			model.oacId = *p.Id
			log.Printf("found OAC for %s with id %s\n", oacc.name, model.oacId)
			found = true
		}
	}
	if found {
		pres.Present(model)
	} else {
		pres.NotFound()
	}
}

func (oacc *OACCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var oacTy driverbottom.Expr
	var sb driverbottom.Expr
	var sp driverbottom.Expr
	for p, v := range oacc.props {
		switch p.Id() {
		case "OriginAccessControlOriginType":
			oacTy = v
		case "SigningBehavior":
			sb = v
		case "SigningProtocol":
			sp = v
		default:
			oacc.tools.Reporter.ReportAtf(p.Loc(), "invalid property for OriginAccessControl: %s", p.Id())
		}
	}

	model := &oacModel{loc: oacc.loc, name: oacc.name, coin: oacc.coin, acType: oacTy, signBehavior: sb, signProt: sp}
	pres.Present(model)
}

func (oacc *OACCreator) UpdateReality() {
	tmp := oacc.tools.Storage.GetCoin(oacc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*oacModel)
		log.Printf("OAC %s already existed for %s\n", found.oacId, found.name)
		return
	}

	desired := oacc.tools.Storage.GetCoin(oacc.coin, corebottom.DETERMINE_DESIRED_MODE).(*oacModel)

	created := &oacModel{loc: desired.loc, name: desired.name, coin: oacc.coin}

	acs, ok1 := oacc.tools.Storage.EvalAsStringer(desired.acType)
	sbs, ok2 := oacc.tools.Storage.EvalAsStringer(desired.signBehavior)
	sps, ok3 := oacc.tools.Storage.EvalAsStringer(desired.signProt)
	if !ok1 || !ok2 || !ok3 {
		panic("not ok")
	}
	ty := types.OriginAccessControlOriginTypes(acs.String())
	sb := types.OriginAccessControlSigningBehaviors(sbs.String())
	sp := types.OriginAccessControlSigningProtocols(sps.String())
	oaccfg := types.OriginAccessControlConfig{Name: &oacc.name, OriginAccessControlOriginType: ty, SigningBehavior: sb, SigningProtocol: sp}
	oac, err := oacc.client.CreateOriginAccessControl(context.TODO(), &cloudfront.CreateOriginAccessControlInput{OriginAccessControlConfig: &oaccfg})
	if err != nil {
		log.Fatalf("failed to create OAC for %s: %v\n", oacc.name, err)
	}
	created.oacId = *oac.OriginAccessControl.Id
	log.Printf("created OAC for %s: %s\n", oacc.name, created.oacId)

	oacc.tools.Storage.Bind(oacc.coin, created)
}

func (oacc *OACCreator) TearDown() {
	tmp := oacc.tools.Storage.GetCoin(oacc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*oacModel)
		log.Printf("you have asked to tear down OAC %s (id: %s) with mode %s\n", oacc.name, found.oacId, oacc.teardown.Mode())
		x, err := oacc.client.GetOriginAccessControl(context.TODO(), &cloudfront.GetOriginAccessControlInput{Id: &found.oacId})
		if err != nil {
			log.Fatalf("could not get OAC %s: %v", found.oacId, err)
		}
		_, err = oacc.client.DeleteOriginAccessControl(context.TODO(), &cloudfront.DeleteOriginAccessControlInput{Id: &found.oacId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete OAC %s: %v", found.oacId, err)
		}
		log.Printf("deleted OAC %s\n", found.oacId)
	} else {
		log.Printf("no OAC existed for %s\n", oacc.name)
	}
}

var _ corebottom.Ensurable = &OACCreator{}

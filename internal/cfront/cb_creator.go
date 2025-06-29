package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type CacheBehaviorCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	// client *cloudfront.Client
}

func (cbc *CacheBehaviorCreator) Loc() *errorsink.Location {
	return cbc.loc
}

func (cbc *CacheBehaviorCreator) ShortDescription() string {
	return "aws.CloudFront.CacheBehavior[" + cbc.name + "]"
}

func (cbc *CacheBehaviorCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.CloudFront.CacheBehavior[")
	iw.AttrsWhere(cbc)
	iw.TextAttr("named", cbc.name)
	if cbc.teardown != nil {
		iw.TextAttr("teardown", cbc.teardown.Mode())
	}
	iw.EndAttrs()
}

func (cbc *CacheBehaviorCreator) CoinId() corebottom.CoinId {
	return cbc.coin
}

func (cbc *CacheBehaviorCreator) Create(pres corebottom.ValuePresenter) {
	var pp driverbottom.Expr
	var rhp driverbottom.Expr
	var cp driverbottom.Expr
	var toid driverbottom.Expr
	for p, v := range cbc.props {
		switch p.Id() {
		case "CachePolicy":
			cp = v
		case "PathPattern":
			pp = v
		case "ResponseHeadersPolicy":
			rhp = v
		case "TargetOriginId":
			toid = v
		default:
			cbc.tools.Reporter.ReportAtf(p.Loc(), "invalid property for OriginAccessControl: %s", p.Id())
		}
	}
	if cp == nil {
		cbc.tools.Reporter.ReportAtf(cbc.loc, "CachePolicy was not defined")
	}
	if rhp == nil {
		cbc.tools.Reporter.ReportAtf(cbc.loc, "ResponseHeadersPolicy was not defined")
	}
	if pp == nil {
		cbc.tools.Reporter.ReportAtf(cbc.loc, "PathPattern was not defined")
	}
	if toid == nil {
		cbc.tools.Reporter.ReportAtf(cbc.loc, "TargetOriginId was not defined")
	}

	ppEval := cbc.tools.Storage.Eval(pp)
	rhpEval := cbc.tools.Storage.Eval(rhp)
	targetOriginId := cbc.tools.Storage.Eval(toid)

	cpId, ok := cbc.tools.Storage.EvalAsStringer(cp)
	if !ok {
		panic("not a string")
	}

	model := &cbModel{pp: ppEval, rhp: rhpEval, targetOriginId: targetOriginId, cpId: cpId}
	// log.Printf("presenting CB %p\n", model)
	pres.Present(model)
}

var _ corebottom.MemoryCoinCreator = &CacheBehaviorCreator{}

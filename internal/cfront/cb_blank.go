package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type CacheBehaviorBlank struct{}

func (b *CacheBehaviorBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
	var pp driverbottom.Expr
	var rhp driverbottom.Expr
	var cp driverbottom.Expr
	var toid driverbottom.Expr
	for p, v := range props {
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
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(p.Loc().Offset, "invalid property for OriginAccessControl: %s", p.Id())
		}
	}
	if cp == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "CachePolicy was not defined")
	}
	if rhp == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "ResponseHeadersPolicy was not defined")
	}
	if pp == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "PathPattern was not defined")
	}
	if toid == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "TargetOriginId was not defined")
	}
	return &CacheBehaviorCreator{tools: tools, teardown: teardown, loc: loc, name: named, cpId: cp, pp: pp, rhp: rhp, toid: toid}
}

func (b *CacheBehaviorBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) any {
	return &CacheBehaviorFinder{tools: tools, loc: loc, name: named}
}

func (b *CacheBehaviorBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *CacheBehaviorBlank) ShortDescription() string {
	return "aws.CloudFront.CacheBehavior[]"
}

func (b *CacheBehaviorBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &CacheBehaviorBlank{}

package cfront

import (
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type CacheBehaviorBlank struct{}

func (b *CacheBehaviorBlank) Mint(tools *pluggable.Tools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr, teardown pluggable.TearDown) any {
	var CacheBehaviorTy pluggable.Expr
	var sb pluggable.Expr
	var sp pluggable.Expr
	for p, v := range props {
		switch p.Id() {
		case "OriginAccessControlOriginType":
			CacheBehaviorTy = v
		case "SigningBehavior":
			sb = v
		case "SigningProtocol":
			sp = v
		default:
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(loc.Offset, "invalid property for OriginAccessControl: %s", p.Id())
		}
	}
	return &CacheBehaviorCreator{tools: tools, loc: loc, name: named, acType: CacheBehaviorTy, signBehavior: sb, signProt: sp, teardown: teardown}
}

func (b *CacheBehaviorBlank) Find(tools *pluggable.Tools, loc *errorsink.Location, named string) any {
	return &CacheBehaviorFinder{tools: tools, loc: loc, name: named}
}

func (b *CacheBehaviorBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *CacheBehaviorBlank) ShortDescription() string {
	return "aws.CloudFront.CacheBehavior[]"
}

func (b *CacheBehaviorBlank) DumpTo(iw pluggable.IndentWriter) {
	panic("not implemented")
}

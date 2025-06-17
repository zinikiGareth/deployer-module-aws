package cfront

import (
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type CachePolicyBlank struct{}

func (b *CachePolicyBlank) Mint(tools *pluggable.Tools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr, teardown pluggable.TearDown) any {
	var CachePolicyTy pluggable.Expr
	var sb pluggable.Expr
	var sp pluggable.Expr
	for p, v := range props {
		switch p.Id() {
		case "OriginAccessControlOriginType":
			CachePolicyTy = v
		case "SigningBehavior":
			sb = v
		case "SigningProtocol":
			sp = v
		default:
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(loc.Offset, "invalid property for OriginAccessControl: %s", p.Id())
		}
	}
	return &CachePolicyCreator{tools: tools, loc: loc, name: named, acType: CachePolicyTy, signBehavior: sb, signProt: sp, teardown: teardown}
}

func (b *CachePolicyBlank) Find(tools *pluggable.Tools, loc *errorsink.Location, named string) any {
	return &CachePolicyFinder{tools: tools, loc: loc, name: named}
}

func (b *CachePolicyBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *CachePolicyBlank) ShortDescription() string {
	return "aws.CloudFront.CachePolicy[]"
}

func (b *CachePolicyBlank) DumpTo(iw pluggable.IndentWriter) {
	panic("not implemented")
}

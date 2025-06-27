package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type CachePolicyBlank struct{}

func (b *CachePolicyBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
	var minttl driverbottom.Expr
	for p, v := range props {
		switch p.Id() {
		case "MinTTL":
			minttl = v
		default:
			tools.Reporter.ReportAtf(loc, "invalid property for OriginAccessControl: %s", p.Id())
		}
	}
	return &CachePolicyCreator{tools: tools, teardown: teardown, loc: loc, name: named, minttl: minttl}
}

func (b *CachePolicyBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) any {
	return &CachePolicyFinder{tools: tools, loc: loc, name: named}
}

func (b *CachePolicyBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *CachePolicyBlank) ShortDescription() string {
	return "aws.CloudFront.CachePolicy[]"
}

func (b *CachePolicyBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &CachePolicyBlank{}

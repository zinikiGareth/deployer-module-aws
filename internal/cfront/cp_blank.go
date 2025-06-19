package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type CachePolicyBlank struct{}

func (b *CachePolicyBlank) Mint(ct *driverbottom.CoreTools, loc *errorsink.Location, named string, props map[driverbottom.Identifier]driverbottom.Expr) any {
	tools := ct.RetrieveOther("coremod").(*corebottom.Tools)
	var minttl driverbottom.Expr
	for p, v := range props {
		switch p.Id() {
		case "MinTTL":
			minttl = v
		default:
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(loc.Offset, "invalid property for OriginAccessControl: %s", p.Id())
		}
	}
	return &CachePolicyCreator{tools: tools, loc: loc, name: named, minttl: minttl}
}

func (b *CachePolicyBlank) Find(ct *driverbottom.CoreTools, loc *errorsink.Location, named string) any {
	return &CachePolicyFinder{tools: ct.RetrieveOther("coremod").(*corebottom.Tools), loc: loc, name: named}
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

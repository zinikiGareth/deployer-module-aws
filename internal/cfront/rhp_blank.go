package cfront

import (
	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type RHPBlank struct{}

func (b *RHPBlank) Mint(ct *pluggable.CoreTools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr) any {
	tools := ct.RetrieveOther("coremod").(*external.Tools)
	var header pluggable.Expr
	var value pluggable.Expr
	for p, v := range props {
		switch p.Id() {
		case "Header":
			header = v
		case "Value":
			value = v
		default:
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(loc.Offset, "invalid property for ResponseHeaderPolicy: %s", p.Id())
		}
	}
	return &RHPCreator{tools: tools, loc: loc, name: named, header: header, value: value}
}

func (b *RHPBlank) Find(ct *pluggable.CoreTools, loc *errorsink.Location, named string) any {
	return &RHPFinder{tools: ct.RetrieveOther("coremod").(*external.Tools), loc: loc, name: named}
}

func (b *RHPBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *RHPBlank) ShortDescription() string {
	return "aws.CloudFront.RHP[]"
}

func (b *RHPBlank) DumpTo(iw pluggable.IndentWriter) {
	panic("not implemented")
}

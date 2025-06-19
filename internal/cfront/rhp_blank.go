package cfront

import (
	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type RHPBlank struct{}

func (b *RHPBlank) Mint(ct *driverbottom.CoreTools, loc *errorsink.Location, named string, props map[driverbottom.Identifier]driverbottom.Expr) any {
	tools := ct.RetrieveOther("coremod").(*external.Tools)
	var header driverbottom.Expr
	var value driverbottom.Expr
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

func (b *RHPBlank) Find(ct *driverbottom.CoreTools, loc *errorsink.Location, named string) any {
	return &RHPFinder{tools: ct.RetrieveOther("coremod").(*external.Tools), loc: loc, name: named}
}

func (b *RHPBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *RHPBlank) ShortDescription() string {
	return "aws.CloudFront.RHP[]"
}

func (b *RHPBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

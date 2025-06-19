package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type RHPBlank struct{}

func (b *RHPBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
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
	return &RHPCreator{tools: tools, teardown: teardown, loc: loc, name: named, header: header, value: value}
}

func (b *RHPBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, named string) any {
	return &RHPFinder{tools: tools, loc: loc, name: named}
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

var _ corebottom.Blank = &RHPBlank{}

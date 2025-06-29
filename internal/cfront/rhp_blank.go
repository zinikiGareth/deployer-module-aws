package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type RHPBlank struct{}

func (b *RHPBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
	return &RHPCreator{tools: tools, teardown: teardown, loc: loc, name: named, coin: id, props: props}
}

func (b *RHPBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) any {
	return &RHPCreator{tools: tools, loc: loc, name: named, coin: id}
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

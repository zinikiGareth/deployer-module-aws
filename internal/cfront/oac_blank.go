package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type OACBlank struct{}

func (b *OACBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
	return &OACCreator{tools: tools, teardown: teardown, loc: loc, coin: id, name: named, props: props}
}

func (b *OACBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) any {
	return &OACCreator{tools: tools, loc: loc, name: named}
}

func (b *OACBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *OACBlank) ShortDescription() string {
	return "aws.CloudFront.OAC[]"
}

func (b *OACBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &OACBlank{}

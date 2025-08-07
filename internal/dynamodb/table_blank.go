package dynamodb

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type TableBlank struct{}

func (b *TableBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &tableCreator{tools: tools, teardown: teardown, loc: loc, coin: id, name: named, props: props}
}

func (b *TableBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr) corebottom.FindCoin {
	return &tableCreator{tools: tools, loc: loc, name: named}
}

func (b *TableBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *TableBlank) ShortDescription() string {
	return "aws.CertificateManager.Certificate[]"
}

func (b *TableBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &TableBlank{}

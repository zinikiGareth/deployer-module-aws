package acm

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type CertificateBlank struct{}

func (b *CertificateBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &certificateCreator{tools: tools, teardown: teardown, loc: loc, coin: id, name: named, props: props}
}

func (b *CertificateBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &certificateCreator{tools: tools, loc: loc, name: named}
}

func (b *CertificateBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *CertificateBlank) ShortDescription() string {
	return "aws.CertificateManager.Certificate[]"
}

func (b *CertificateBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &CertificateBlank{}

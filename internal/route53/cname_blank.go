package route53

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type CNAMEBlank struct{}

func (b *CNAMEBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &cnameCreator{tools: tools, loc: loc, name: named, coin: id, props: props}
}

func (b *CNAMEBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &cnameCreator{tools: tools, loc: loc, name: named, coin: id}
}

func (b *CNAMEBlank) ShortDescription() string {
	return "test.S3.Bucket[]"
}

var _ corebottom.Blank = &CNAMEBlank{}

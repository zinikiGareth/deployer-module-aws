package lambda

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type AliasBlank struct{}

func (b *AliasBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	panic("can't mint an alias - use publishVersion")
}

func (b *AliasBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr) corebottom.FindCoin {
	return &aliasFinder{tools: tools, loc: loc, name: named, coin: id, props: props}
}

func (b *AliasBlank) ShortDescription() string {
	return "aws.Lambda.Alias[]"
}

var _ corebottom.Blank = &AliasBlank{}

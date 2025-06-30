package route53

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type ALIASBlank struct{}

func (b *ALIASBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &aliasCreator{tools: tools, teardown: teardown, coin: id, loc: loc, name: named, props: props}
}

func (b *ALIASBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &aliasCreator{tools: tools, loc: loc, coin: id, name: named}
}

func (b *ALIASBlank) ShortDescription() string {
	return "test.S3.Bucket[]"
}

var _ corebottom.Blank = &ALIASBlank{}

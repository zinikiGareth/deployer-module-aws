package lambda

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type FunctionBlank struct{}

func (b *FunctionBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &lambdaCreator{tools: tools, teardown: teardown, loc: loc, name: named, coin: id, props: props}
}

func (b *FunctionBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr) corebottom.FindCoin {
	return &lambdaCreator{tools: tools, loc: loc, name: named, coin: id}
}

func (b *FunctionBlank) ShortDescription() string {
	return "aws.Lambda.Function[]"
}

var _ corebottom.Blank = &FunctionBlank{}

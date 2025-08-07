package gatewayV2

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type StageBlank struct{}

func (b *StageBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &stageCreator{tools: tools, teardown: teardown, loc: loc, name: named, coin: id, props: props}
}

func (b *StageBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr) corebottom.FindCoin {
	return &stageCreator{tools: tools, loc: loc, name: named, coin: id}
}

func (b *StageBlank) ShortDescription() string {
	return "aws.ApiGatewayV2.Stage[]"
}

var _ corebottom.Blank = &StageBlank{}

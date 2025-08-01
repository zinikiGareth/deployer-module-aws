package gatewayV2

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type IntegrationBlank struct{}

func (b *IntegrationBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &integrationCreator{tools: tools, teardown: teardown, loc: loc, name: named, coin: id, props: props}
}

func (b *IntegrationBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &integrationCreator{tools: tools, loc: loc, name: named, coin: id}
}

func (b *IntegrationBlank) ShortDescription() string {
	return "aws.ApiGatewayV2.Integration[]"
}

var _ corebottom.Blank = &IntegrationBlank{}

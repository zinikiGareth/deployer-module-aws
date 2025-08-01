package gatewayV2

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type RouteBlank struct{}

func (b *RouteBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &routeCreator{tools: tools, teardown: teardown, loc: loc, path: named, coin: id, props: props}
}

func (b *RouteBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &routeCreator{tools: tools, loc: loc, path: named, coin: id}
}

func (b *RouteBlank) ShortDescription() string {
	return "aws.ApiGatewayV2.Route[]"
}

var _ corebottom.Blank = &RouteBlank{}

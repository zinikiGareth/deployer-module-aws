package gatewayV2

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type VPCLinkBlank struct{}

func (b *VPCLinkBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &vpcLinkCreator{tools: tools, teardown: teardown, loc: loc, name: named, coin: id, props: props}
}

func (b *VPCLinkBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &vpcLinkCreator{tools: tools, loc: loc, name: named, coin: id}
}

func (b *VPCLinkBlank) ShortDescription() string {
	return "aws.ApiGatewayV2.VPCLink[]"
}

var _ corebottom.Blank = &VPCLinkBlank{}

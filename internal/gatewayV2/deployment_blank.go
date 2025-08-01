package gatewayV2

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type DeploymentBlank struct{}

func (b *DeploymentBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &deploymentCreator{tools: tools, teardown: teardown, loc: loc, name: named, coin: id, props: props}
}

func (b *DeploymentBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &deploymentCreator{tools: tools, loc: loc, name: named, coin: id}
}

func (b *DeploymentBlank) ShortDescription() string {
	return "aws.ApiGatewayV2.Deployment[]"
}

var _ corebottom.Blank = &DeploymentBlank{}

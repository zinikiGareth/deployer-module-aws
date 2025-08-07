package neptune

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type InstanceBlank struct{}

func (b *InstanceBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &instanceCreator{tools: tools, teardown: teardown, loc: loc, coin: id, name: named, props: props}
}

func (b *InstanceBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr) corebottom.FindCoin {
	return &instanceCreator{tools: tools, loc: loc, name: named}
}

func (b *InstanceBlank) ShortDescription() string {
	return "aws.Neptune.instanceBlank[]"
}

var _ corebottom.Blank = &InstanceBlank{}

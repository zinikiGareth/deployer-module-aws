package neptune

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type ClusterBlank struct{}

func (b *ClusterBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &clusterCreator{tools: tools, teardown: teardown, loc: loc, coin: id, name: named, props: props}
}

func (b *ClusterBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &clusterCreator{tools: tools, loc: loc, name: named}
}

func (b *ClusterBlank) ShortDescription() string {
	return "aws.Neptune.ClusterBlank[]"
}

var _ corebottom.Blank = &ClusterBlank{}

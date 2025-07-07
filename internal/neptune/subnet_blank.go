package neptune

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type SubnetBlank struct{}

func (b *SubnetBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	tools.Reporter.ReportAtf(loc, "cannot create subnet groups at this time")
	return nil
	// return &subnetCreator{tools: tools, teardown: teardown, loc: loc, coin: id, name: named, props: props}
}

func (b *SubnetBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &subnetCreator{tools: tools, loc: loc, name: named}
}

func (b *SubnetBlank) ShortDescription() string {
	return "aws.Neptune.SubnetBlank[]"
}

var _ corebottom.Blank = &SubnetBlank{}

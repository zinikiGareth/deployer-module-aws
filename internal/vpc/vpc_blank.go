package vpc

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type VPCBlank struct{}

func (b *VPCBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	tools.Reporter.ReportAtf(loc, "cannot create domain names automatically; use find")
	return nil
}

func (b *VPCBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr) corebottom.FindCoin {
	return &vpcFinder{tools: tools, loc: loc, name: named}
}

func (b *VPCBlank) ShortDescription() string {
	return "aws.VPC.VPC"
}

var _ corebottom.Blank = &VPCBlank{}

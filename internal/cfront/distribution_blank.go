package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type DistributionBlank struct{}

func (b *DistributionBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
	return &distributionCreator{tools: tools, teardown: teardown, loc: loc, name: named, coin: id, props: props}
}

func (b *DistributionBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) any {
	return &distributionCreator{tools: tools, loc: loc, name: named, coin: id}
}

func (b *DistributionBlank) ShortDescription() string {
	return "aws.CloudFront.Distribution[]"
}

var _ corebottom.Blank = &DistributionBlank{}

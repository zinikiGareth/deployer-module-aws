package cfront

import (
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type DistributionBlank struct{}

func (b *DistributionBlank) Mint(tools *pluggable.Tools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr, teardown pluggable.TearDown) any {
	return &distributionCreator{tools: tools, loc: loc, name: named, props: props, teardown: teardown}
}

func (b *DistributionBlank) Find(tools *pluggable.Tools, loc *errorsink.Location, named string) any {
	return &distributionFinder{tools: tools, loc: loc, name: named}
}

func (b *DistributionBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *DistributionBlank) ShortDescription() string {
	return "aws.CloudFront.Distribution[]"
}

func (b *DistributionBlank) DumpTo(iw pluggable.IndentWriter) {
	panic("not implemented")
}

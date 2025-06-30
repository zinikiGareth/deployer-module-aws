package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type CacheBehaviorBlank struct{}

func (b *CacheBehaviorBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr) corebottom.MemoryCoinCreator {
	return &CacheBehaviorCreator{tools: tools, loc: loc, name: named, coin: id, props: props}
}

func (b *CacheBehaviorBlank) ShortDescription() string {
	return "aws.CloudFront.CacheBehavior[]"
}

var _ corebottom.MemoryCoin = &CacheBehaviorBlank{}

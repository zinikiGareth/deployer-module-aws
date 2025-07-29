package iam

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type PolicyBlank struct{}

func (b *PolicyBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &policyCreator{tools: tools, loc: loc, name: named, coin: id, props: props, teardown: teardown}
}

func (b *PolicyBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &policyCreator{tools: tools, loc: loc, name: named, coin: id}
}

func (b *PolicyBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *PolicyBlank) ShortDescription() string {
	return "test.IAM.Policy[]"
}

func (b *PolicyBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &PolicyBlank{}

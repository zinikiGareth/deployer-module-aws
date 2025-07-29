package iam

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type RoleBlank struct{}

func (b *RoleBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	return &roleCreator{tools: tools, teardown: teardown, loc: loc, name: named, coin: id}
}

func (b *RoleBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &roleCreator{tools: tools, loc: loc, name: named, coin: id}
}

func (b *RoleBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *RoleBlank) ShortDescription() string {
	return "test.IAM.Role[]"
}

func (b *RoleBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &RoleBlank{}

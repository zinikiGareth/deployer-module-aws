package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type OACBlank struct{}

func (b *OACBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
	var oacTy driverbottom.Expr
	var sb driverbottom.Expr
	var sp driverbottom.Expr
	for p, v := range props {
		switch p.Id() {
		case "OriginAccessControlOriginType":
			oacTy = v
		case "SigningBehavior":
			sb = v
		case "SigningProtocol":
			sp = v
		default:
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(loc.Offset, "invalid property for OriginAccessControl: %s", p.Id())
		}
	}
	return &OACCreator{tools: tools, teardown: teardown, loc: loc, name: named, acType: oacTy, signBehavior: sb, signProt: sp}
}

func (b *OACBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) any {
	return &OACFinder{tools: tools, loc: loc, name: named}
}

func (b *OACBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *OACBlank) ShortDescription() string {
	return "aws.CloudFront.OAC[]"
}

func (b *OACBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &OACBlank{}

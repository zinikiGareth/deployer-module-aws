package cfront

import (
	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type OACBlank struct{}

func (b *OACBlank) Mint(tools *external.Tools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr) any {
	var oacTy pluggable.Expr
	var sb pluggable.Expr
	var sp pluggable.Expr
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
	return &OACCreator{tools: tools, loc: loc, name: named, acType: oacTy, signBehavior: sb, signProt: sp}
}

func (b *OACBlank) Find(tools *external.Tools, loc *errorsink.Location, named string) any {
	return &OACFinder{tools: tools, loc: loc, name: named}
}

func (b *OACBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *OACBlank) ShortDescription() string {
	return "aws.CloudFront.OAC[]"
}

func (b *OACBlank) DumpTo(iw pluggable.IndentWriter) {
	panic("not implemented")
}

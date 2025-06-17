package cfront

import (
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type DistributionBlank struct{}

func (b *DistributionBlank) Mint(tools *pluggable.Tools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr, teardown pluggable.TearDown) any {
	var cert pluggable.Expr
	var domain pluggable.Expr
	var oac pluggable.Expr
	for p, v := range props {
		switch p.Id() {
		case "Certificate":
			cert = v
		case "Domain":
			domain = v
		case "OriginAccessControl":
			oac = v
		default:
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(loc.Offset, "invalid property for Distribution: %s", p.Id())
		}
	}
	return &distributionCreator{tools: tools, loc: loc, name: named, oac: oac, domain: domain, viewerCert: cert, teardown: teardown}
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

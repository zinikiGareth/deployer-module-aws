package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type DistributionBlank struct{}

func (b *DistributionBlank) Mint(ct *driverbottom.CoreTools, loc *errorsink.Location, named string, props map[driverbottom.Identifier]driverbottom.Expr) any {
	tools := ct.RetrieveOther("coremod").(*corebottom.Tools)
	var cert driverbottom.Expr
	var domain driverbottom.Expr
	var oac driverbottom.Expr
	var cbs driverbottom.Expr
	var cp driverbottom.Expr
	var comment driverbottom.Expr
	var src driverbottom.Expr
	var toid driverbottom.Expr
	for p, v := range props {
		switch p.Id() {
		case "Certificate":
			cert = v
		case "OriginDNS":
			src = v
		case "Comment":
			comment = v
		case "Domain":
			domain = v
		case "OriginAccessControl":
			oac = v
		case "CacheBehaviors":
			cbs = v
		case "CachePolicy":
			cp = v
		case "TargetOriginId":
			toid = v
		default:
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(loc.Offset, "invalid property for Distribution: %s", p.Id())
		}
	}
	if comment == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "Comment was not defined")
	}
	if src == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "OriginDNS was not defined")
	}
	if toid == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "TargetOriginId was not defined")
	}
	return &distributionCreator{tools: tools, loc: loc, name: named, comment: comment, origindns: src, oac: oac, behaviors: cbs, cachePolicy: cp, domain: domain, viewerCert: cert, toid: toid}
}

func (b *DistributionBlank) Find(ct *driverbottom.CoreTools, loc *errorsink.Location, named string) any {
	return &distributionFinder{tools: ct.RetrieveOther("coremod").(*corebottom.Tools), loc: loc, name: named}
}

func (b *DistributionBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *DistributionBlank) ShortDescription() string {
	return "aws.CloudFront.Distribution[]"
}

func (b *DistributionBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

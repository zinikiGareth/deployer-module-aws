package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type DistributionBlank struct{}

func (b *DistributionBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
	var cert driverbottom.Expr
	var domain driverbottom.List
	var oac driverbottom.Expr
	var cbs driverbottom.List
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
			le, isList := v.(driverbottom.List)
			if isList {
				domain = le
			} else {
				domain = drivertop.NewListExpr([]driverbottom.Expr{v})
			}
		case "OriginAccessControl":
			oac = v
		case "CacheBehaviors":
			le, isList := v.(driverbottom.List)
			if isList {
				cbs = le
			} else {
				cbs = drivertop.NewListExpr([]driverbottom.Expr{v})
			}
		case "CachePolicy":
			cp = v
		case "TargetOriginId":
			toid = v
		default:
			tools.Reporter.ReportAtf(loc, "invalid property for Distribution: %s", p.Id())
		}
	}
	if comment == nil {
		tools.Reporter.ReportAtf(loc, "Comment was not defined")
	}
	if src == nil {
		tools.Reporter.ReportAtf(loc, "OriginDNS was not defined")
	}
	if toid == nil {
		tools.Reporter.ReportAtf(loc, "TargetOriginId was not defined")
	}
	return &distributionCreator{tools: tools, teardown: teardown, loc: loc, name: named, comment: comment, origindns: src, oac: oac, behaviors: cbs, cachePolicy: cp, domains: domain, viewerCert: cert, toid: toid}
}

func (b *DistributionBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) any {
	return &distributionFinder{tools: tools, loc: loc, name: named}
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

var _ corebottom.Blank = &DistributionBlank{}

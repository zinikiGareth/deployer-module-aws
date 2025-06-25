package route53

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type CNAMEBlank struct{}

func (b *CNAMEBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
	var pointsTo driverbottom.Expr
	var zone driverbottom.Expr
	seenErr := false
	for p, v := range props {
		switch p.Id() {
		case "PointsTo":
			pointsTo = v
		case "Zone":
			zone = v
		default:
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(loc.Offset, "invalid property for IAM policy: %s", p.Id())
		}
	}
	if !seenErr && zone == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "no Zone property was specified for %s", named)
	}
	if !seenErr && pointsTo == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "no PointsTo property was specified for %s", named)
	}
	return &cnameCreator{tools: tools, teardown: teardown, loc: loc, name: named, pointsTo: pointsTo, zone: zone}
}

func (b *CNAMEBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) any {
	panic("not implemented")
}

func (b *CNAMEBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *CNAMEBlank) ShortDescription() string {
	return "test.S3.Bucket[]"
}

func (b *CNAMEBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &CNAMEBlank{}

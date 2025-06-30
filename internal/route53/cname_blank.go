package route53

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type CNAMEBlank struct{}

func (b *CNAMEBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
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
			tools.Reporter.ReportAtf(loc, "invalid property for IAM policy: %s", p.Id())
		}
	}
	if !seenErr && zone == nil {
		tools.Reporter.ReportAtf(loc, "no Zone property was specified for %s", named)
	}
	if !seenErr && pointsTo == nil {
		tools.Reporter.ReportAtf(loc, "no PointsTo property was specified for %s", named)
	}
	return &cnameCreator{tools: tools, teardown: teardown, loc: loc, name: named, coin: id, props: props}
}

func (b *CNAMEBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &cnameCreator{tools: tools, loc: loc, name: named, coin: id}
}

func (b *CNAMEBlank) ShortDescription() string {
	return "test.S3.Bucket[]"
}

var _ corebottom.Blank = &CNAMEBlank{}

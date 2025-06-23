package route53

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type ALIASBlank struct{}

func (b *ALIASBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
	var pointsTo driverbottom.Expr
	var updZone driverbottom.Expr
	var aliasZone driverbottom.Expr
	seenErr := false
	for p, v := range props {
		switch p.Id() {
		case "PointsTo":
			pointsTo = v
		case "UpdateZone":
			updZone = v
		case "AliasZone":
			aliasZone = v
		default:
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(loc.Offset, "invalid property for IAM policy: %s", p.Id())
		}
	}
	if !seenErr && updZone == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "no UpdateZone property was specified for %s", named)
	}
	if !seenErr && aliasZone == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "no AliasZone property was specified for %s", named)
	}
	if !seenErr && pointsTo == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "no PointsTo property was specified for %s", named)
	}
	return &aliasCreator{tools: tools, teardown: teardown, loc: loc, name: named, pointsTo: pointsTo, updateZone: updZone, aliasZone: aliasZone}
}

func (b *ALIASBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, named string) any {
	panic("not implemented")
}

func (b *ALIASBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *ALIASBlank) ShortDescription() string {
	return "test.S3.Bucket[]"
}

func (b *ALIASBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &ALIASBlank{}

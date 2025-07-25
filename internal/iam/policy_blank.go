package iam

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type PolicyBlank struct{}

func (b *PolicyBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	var policy driverbottom.Expr
	seenErr := false
	for p, v := range props {
		switch p.Id() {
		case "Policy":
			policy = v
		default:
			tools.Reporter.ReportAtf(loc, "invalid property for IAM policy: %s", p.Id())
		}
	}
	if !seenErr && policy == nil {
		tools.Reporter.ReportAtf(loc, "no Policy property was specified for %s", named)
	}
	return &policyCreator{tools: tools, teardown: teardown, loc: loc, name: named, coin: id, policy: policy}
}

func (b *PolicyBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &policyFinder{tools: tools, loc: loc, name: named, coin: id}
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

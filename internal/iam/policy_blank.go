package iam

import (
	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type PolicyBlank struct{}

func (b *PolicyBlank) Mint(ct *pluggable.CoreTools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr) any {
	tools := ct.RetrieveOther("coremod").(*external.Tools)
	var policy pluggable.Expr
	seenErr := false
	for p, v := range props {
		switch p.Id() {
		case "Policy":
			policy = v
		default:
			tools.Reporter.At(p.Loc().Line)
			tools.Reporter.Reportf(loc.Offset, "invalid property for IAM policy: %s", p.Id())
		}
	}
	if !seenErr && policy == nil {
		tools.Reporter.At(loc.Line)
		tools.Reporter.Reportf(loc.Offset, "no Policy property was specified for %s", named)
	}
	return &policyCreator{tools: tools, loc: loc, name: named, policy: policy}
}

func (b *PolicyBlank) Find(ct *pluggable.CoreTools, loc *errorsink.Location, named string) any {
	return &policyFinder{tools: ct.RetrieveOther("coremod").(*external.Tools), loc: loc, name: named}
}

func (b *PolicyBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *PolicyBlank) ShortDescription() string {
	return "test.IAM.Policy[]"
}

func (b *PolicyBlank) DumpTo(iw pluggable.IndentWriter) {
	panic("not implemented")
}

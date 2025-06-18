package route53

import (
	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type DomainNameBlank struct{}

func (b *DomainNameBlank) Mint(tools *external.Tools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr, teardown external.TearDown) any {
	tools.Reporter.At(loc.Line)
	tools.Reporter.Reportf(loc.Offset, "cannot create domain names automatically; use find")
	return nil
}

func (b *DomainNameBlank) Find(tools *external.Tools, loc *errorsink.Location, named string) any {
	return &domainNameFinder{tools: tools, loc: loc, name: named}
}

func (b *DomainNameBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *DomainNameBlank) ShortDescription() string {
	return "test.S3.Bucket[]"
}

func (b *DomainNameBlank) DumpTo(iw pluggable.IndentWriter) {
	panic("not implemented")
}

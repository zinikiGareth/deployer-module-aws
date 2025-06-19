package route53

import (
	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type DomainNameBlank struct{}

func (b *DomainNameBlank) Mint(ct *driverbottom.CoreTools, loc *errorsink.Location, named string, props map[driverbottom.Identifier]driverbottom.Expr) any {
	tools := ct.RetrieveOther("coremod").(*external.Tools)
	tools.Reporter.At(loc.Line)
	tools.Reporter.Reportf(loc.Offset, "cannot create domain names automatically; use find")
	return nil
}

func (b *DomainNameBlank) Find(ct *driverbottom.CoreTools, loc *errorsink.Location, named string) any {
	return &domainNameFinder{tools: ct.RetrieveOther("coremod").(*external.Tools), loc: loc, name: named}
}

func (b *DomainNameBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *DomainNameBlank) ShortDescription() string {
	return "test.S3.Bucket[]"
}

func (b *DomainNameBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

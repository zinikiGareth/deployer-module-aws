package route53

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type DomainNameBlank struct{}

func (b *DomainNameBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) corebottom.Ensurable {
	tools.Reporter.ReportAtf(loc, "cannot create domain names automatically; use find")
	return nil
}

func (b *DomainNameBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) corebottom.FindCoin {
	return &domainNameFinder{tools: tools, loc: loc, name: named}
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

var _ corebottom.Blank = &DomainNameBlank{}

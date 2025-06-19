package acm

import (
	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type CertificateBlank struct{}

func (b *CertificateBlank) Mint(ct *pluggable.CoreTools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr) any {
	return &certificateCreator{tools: ct.RetrieveOther("coremod").(*external.Tools), loc: loc, name: named, props: props}
}

func (b *CertificateBlank) Find(ct *pluggable.CoreTools, loc *errorsink.Location, named string) any {
	return &certificateFinder{tools: ct.RetrieveOther("coremod").(*external.Tools), loc: loc, name: named}
}

func (b *CertificateBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *CertificateBlank) ShortDescription() string {
	return "aws.CertificateManager.Certificate[]"
}

func (b *CertificateBlank) DumpTo(iw pluggable.IndentWriter) {
	panic("not implemented")
}

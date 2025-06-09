package acm

import (
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type CertificateBlank struct{}

func (b *CertificateBlank) Mint(tools *pluggable.Tools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr, teardown pluggable.TearDown) any {
	return &certificateCreator{tools: tools, loc: loc, name: named, props: props, teardown: teardown}
}

func (b *CertificateBlank) Find(tools *pluggable.Tools, loc *errorsink.Location, named string) any {
	return &certificateFinder{tools: tools, loc: loc, name: named}
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

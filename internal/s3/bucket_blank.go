package s3

import (
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type BucketBlank struct{}

func (b *BucketBlank) Mint(tools *pluggable.Tools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr, teardown pluggable.TearDown) any {
	return &bucketCreator{tools: tools, loc: loc, name: named, teardown: teardown}
}

func (b *BucketBlank) Find(tools *pluggable.Tools, loc *errorsink.Location, named string) any {
	return &bucketFinder{tools: tools, loc: loc, name: named}
}

func (b *BucketBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *BucketBlank) ShortDescription() string {
	return "test.S3.Bucket[]"
}

func (b *BucketBlank) DumpTo(iw pluggable.IndentWriter) {
	panic("not implemented")
}

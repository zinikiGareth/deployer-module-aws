package s3

import (
	"ziniki.org/deployer/deployer/pkg/errors"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type BucketBlank struct{}

func (b *BucketBlank) Mint(tools *pluggable.Tools, loc *errors.Location, named string) any {
	return &bucketCreator{tools: tools, loc: loc, name: named}
}

func (b *BucketBlank) Loc() *errors.Location {
	panic("not implemented")
}

func (b *BucketBlank) ShortDescription() string {
	return "test.S3.Bucket[]"
}

func (b *BucketBlank) DumpTo(iw pluggable.IndentWriter) {
	panic("not implemented")
}

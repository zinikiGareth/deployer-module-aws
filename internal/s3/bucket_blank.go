package s3

import (
	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type BucketBlank struct{}

func (b *BucketBlank) Mint(ct *pluggable.CoreTools, loc *errorsink.Location, named string, props map[pluggable.Identifier]pluggable.Expr) any {
	tools := ct.RetrieveOther("coremod").(*external.Tools)
	return &bucketCreator{tools: tools, loc: loc, name: named, region: "us-east-1"}
}

func (b *BucketBlank) Find(ct *pluggable.CoreTools, loc *errorsink.Location, named string) any {
	return &bucketFinder{tools: ct.RetrieveOther("coremod").(*external.Tools), loc: loc, name: named}
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

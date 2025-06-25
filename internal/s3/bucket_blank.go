package s3

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type BucketBlank struct{}

func (b *BucketBlank) Mint(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string, props map[driverbottom.Identifier]driverbottom.Expr, teardown corebottom.TearDown) any {
	return &bucketCreator{tools: tools, teardown: teardown, loc: loc, name: named, props: props}
}

func (b *BucketBlank) Find(tools *corebottom.Tools, loc *errorsink.Location, id corebottom.CoinId, named string) any {
	return &bucketFinder{tools: tools, loc: loc, name: named}
}

func (b *BucketBlank) Loc() *errorsink.Location {
	panic("not implemented")
}

func (b *BucketBlank) ShortDescription() string {
	return "test.S3.Bucket[]"
}

func (b *BucketBlank) DumpTo(iw driverbottom.IndentWriter) {
	panic("not implemented")
}

var _ corebottom.Blank = &BucketBlank{}

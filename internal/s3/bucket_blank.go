package s3

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type BucketBlank struct{}

func (b *BucketBlank) Mint(ct *driverbottom.CoreTools, loc *errorsink.Location, named string, props map[driverbottom.Identifier]driverbottom.Expr) any {
	tools := ct.RetrieveOther("coremod").(*corebottom.Tools)
	return &bucketCreator{tools: tools, loc: loc, name: named, region: "us-east-1"}
}

func (b *BucketBlank) Find(ct *driverbottom.CoreTools, loc *errorsink.Location, named string) any {
	return &bucketFinder{tools: ct.RetrieveOther("coremod").(*corebottom.Tools), loc: loc, name: named}
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

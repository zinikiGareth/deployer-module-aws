package s3

import (
	"context"
	e "errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type bucketCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	coin     driverbottom.Holder
	teardown corebottom.TearDown
	name     string // name is here (as well as?) the model because it's core to who we are
	props    map[driverbottom.Identifier]driverbottom.Expr

	client *s3.Client
	// alreadyExists bool
	// model         *bucketModel
	// cloud *BucketCloud
}

func (b *bucketCreator) Loc() *errorsink.Location {
	return b.loc
}

func (b *bucketCreator) ShortDescription() string {
	return "aws.s3.BucketCreator[" + b.name + "]"
}

func (b *bucketCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.s3.BucketCreator[")
	iw.AttrsWhere(b)
	iw.TextAttr("named", b.name)
	iw.IndPrintf("properties\n")
	iw.Indent()
	for i, e := range b.props {
		iw.NestedAttr(i.Id(), e)
	}
	iw.UnIndent()
	iw.EndAttrs()
}

func (b *bucketCreator) CoinId() corebottom.CoinId {
	return b.coin
}

func (b *bucketCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := b.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	b.client = awsEnv.S3Client()
	_, err := b.client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(b.name),
	})
	if err != nil {
		var api smithy.APIError
		if e.As(err, &api) {
			// log.Printf("code: %s", api.ErrorCode())
			if api.ErrorCode() == "NotFound" {
				log.Printf("bucket does not exist: %s", b.name)
				pres.NotFound()
			} else {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	} else {
		log.Printf("bucket exists: %s", b.name)
		model := &bucketModel{loc: b.loc, storage: b.tools.Storage, id: b.coin, client: b.client, name: b.name}
		pres.Present(model)
	}
}

func (b *bucketCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	region, _ := utils.AsStringer("us-east-1")
	model := &bucketModel{loc: b.loc, storage: b.tools.Storage, id: b.coin, client: b.client, name: b.name}
	// TODO: should this be an earlier phase?
	for i, e := range b.props {
		v := b.tools.Storage.Eval(e)
		switch i.Id() {
		case "Region":
			var ok bool
			region, ok = v.(fmt.Stringer)
			if !ok {
				b.tools.Reporter.ReportAtf(e.Loc(), "must be a string value")
				return
			}
		}
	}
	model.region = region

	pres.Present(model)
}

func (b *bucketCreator) UpdateReality() {
	tmp := b.tools.Storage.GetCoin(b.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*bucketModel)
		log.Printf("bucket %s already existed\n", found.name)
		return
	}

	bucket := CreateBucket(b.client, b.name)
	if bucket != nil {
		log.Printf("created bucket %s\n", *bucket.Location)
	}
}

func (b *bucketCreator) TearDown() {
	tmp := b.tools.Storage.GetCoin(b.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp == nil {
		log.Printf("bucket %s does not exist\n", b.name)
		return
	}
	log.Printf("you have asked to tear down bucket %s %s\n", b.name, b.teardown.Mode())
	switch b.teardown.Mode() {
	case "preserve":
		log.Printf("not deleting bucket %s because teardown mode is 'preserve'", b.name)
		// case "empty" seems like it might be a reasonable option
	case "delete":
		log.Printf("deleting bucket %s with teardown mode 'delete'", b.name)
		EmptyBucket(b.client, b.name)
		DeleteBucket(b.client, b.name)
	default:
		log.Printf("cannot handle teardown mode '%s' for bucket %s", b.teardown.Mode(), b.name)
	}
}

func (b *bucketCreator) String() string {
	return fmt.Sprintf("EnsureBucket[%s:%s]", "" /* eb.env.Region */, b.name)
}

var _ corebottom.Ensurable = &bucketCreator{}

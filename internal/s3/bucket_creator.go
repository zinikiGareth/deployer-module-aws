package s3

import (
	"context"
	e "errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/files"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type bucketCreator struct {
	tools *pluggable.Tools

	loc      *errorsink.Location
	name     string
	teardown pluggable.TearDown

	client        *s3.Client
	alreadyExists bool
	// cloud *BucketCloud
}

func (b *bucketCreator) Loc() *errorsink.Location {
	return b.loc
}

func (b *bucketCreator) ShortDescription() string {
	return "aws.s3.Bucket[" + b.name + "]"
}

func (b *bucketCreator) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.s3.Bucket[")
	iw.AttrsWhere(b)
	iw.TextAttr("named", b.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (b *bucketCreator) BuildModel(pres pluggable.ValuePresenter) {
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
			} else {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	} else {
		log.Printf("bucket exists: %s", b.name)
		b.alreadyExists = true
	}

	// TODO: do we need to capture something here?
	pres.Present(b)
}

func (b *bucketCreator) UpdateReality() {
	if b.alreadyExists {
		log.Printf("bucket %s already existed\n", b.name)
		return
	}

	bucket := CreateBucket(b.client, b.name)
	if bucket != nil {
		log.Printf("created bucket %s\n", *bucket.Location)
	}
}

func (b *bucketCreator) TearDown() {
	if !b.alreadyExists {
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

func (b *bucketCreator) ObtainDest() files.FileDest {
	return NewBucketTransfer(b.client, b.name)
}

func (b *bucketCreator) ObtainMethod(name string) pluggable.Method {
	switch name {
	case "allResources":
		return &allResourcesMethod{}
	}
	return nil
}

func (b *bucketCreator) String() string {
	return fmt.Sprintf("EnsureBucket[%s:%s]", "" /* eb.env.Region */, b.name)
}

type allResourcesMethod struct {
}

func (a *allResourcesMethod) Invoke(s pluggable.RuntimeStorage, on pluggable.Expr, args []pluggable.Expr) any {
	e := on.Eval(s)
	bucket, ok := e.(*bucketCreator)
	if !ok {
		panic(fmt.Sprintf("allResources can only be called on a bucket, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	return fmt.Sprintf("arn:aws:s3:::%s/*", bucket.name)
}

var _ pluggable.HasMethods = &bucketCreator{}
var _ pluggable.Method = &allResourcesMethod{}

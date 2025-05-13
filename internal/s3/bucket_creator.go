package s3

import (
	"context"
	e "errors"
	"fmt"
	"log"
	"time"

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

	client *s3.Client
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
func (b *bucketCreator) Prepare(pres pluggable.ValuePresenter) {
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
				return
			}
		}
		log.Fatal(err)
	}

	log.Printf("bucket exists: %s", b.name)
	/*

		tmp = b.tools.Recall.ObtainDriver("testhelpers.TestStepLogger")
		testLogger, ok := tmp.(testhelpers.TestStepLogger)
		if !ok {
			panic("could not cast logger to TestStepLogger")
		}

		b.env = testAwsEnv
		testLogger.Log("ensuring bucket exists action %s\n", b.String())
		pres.Present(b)
	*/
}

func (b *bucketCreator) Execute() {
	// TODO: need to consider cases ...
	bucket, err := b.client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(b.name),
	})
	if err != nil {
		log.Fatalf("error creating bucket %s: %v\n", b.name, err)
	} else {
		err = s3.NewBucketExistsWaiter(b.client).Wait(
			context.TODO(), &s3.HeadBucketInput{Bucket: aws.String(b.name)},
			time.Minute,
		)
		if err != nil {
			log.Printf("Failed attempt to wait for bucket %s to exist: %v.\n", b.name, err)
		}
	}
	// b.cloud = bucket.
	log.Printf("created bucket %s\n", *bucket.Location)

	/*
		tmp := eb.tools.Recall.ObtainDriver("testhelpers.TestStepLogger")
		testLogger, ok := tmp.(testhelpers.TestStepLogger)
		if !ok {
			panic("could not cast logger to TestStepLogger")
		}

		b := eb.env.FindBucket(eb.name)
		if b != nil {
			testLogger.Log("the bucket %s in region %s already exists\n", eb.name, eb.env.Region)
		} else {
			testLogger.Log("we need to create a bucket called %s in region %s\n", eb.name, eb.env.Region)
			// TODO: we should also handle all the properties we have stored somewhere ...
			b = eb.env.CreateBucket(eb.name)
		}

		eb.cloud = b
	*/
}

func (b *bucketCreator) TearDown() {
	log.Printf("you have asked to tear down bucket %s %s\n", b.name, b.teardown.Mode())
	switch b.teardown.Mode() {
	case "preserve":
		log.Printf("not deleting bucket %s because teardown mode is 'preserve'", b.name)
	case "delete":
		log.Printf("deleting bucket %s with teardown mode 'delete'", b.name)
		out, err := b.client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
			Bucket: aws.String(b.name),
		})
		if err != nil {
			log.Fatalf("error deleting bucket %s: %v\n", b.name, err)
		} else {
			log.Printf("Deleted bucket %s: %v\n", b.name, out)
		}
	default:
		log.Printf("cannot handle teardown mode '%s' for bucket %s", b.teardown.Mode(), b.name)
	}
}

func (b *bucketCreator) ObtainDest() files.FileDest {
	return nil // eb.cloud
}

func (b *bucketCreator) String() string {
	return fmt.Sprintf("EnsureBucket[%s:%s]", "" /* eb.env.Region */, b.name)
}

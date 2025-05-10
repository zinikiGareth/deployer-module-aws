package s3

import (
	"context"
	e "errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/files"
	"ziniki.org/deployer/deployer/pkg/errors"
	"ziniki.org/deployer/deployer/pkg/pluggable"
)

type bucketCreator struct {
	tools *pluggable.Tools

	loc  *errors.Location
	name string

	// env   *TestAwsEnv
	// cloud *BucketCloud
}

func (b *bucketCreator) Loc() *errors.Location {
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
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	client := s3.NewFromConfig(cfg)
	_, err = client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
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
		tmp := b.tools.Recall.ObtainDriver("testS3.TestAwsEnv")
		testAwsEnv, ok := tmp.(*TestAwsEnv)
		if !ok {
			panic("could not cast env to TestAwsEnv")
		}

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

func (eb *bucketCreator) Execute() {
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

func (eb *bucketCreator) ObtainDest() files.FileDest {
	return nil // eb.cloud
}

func (eb *bucketCreator) String() string {
	return fmt.Sprintf("EnsureBucket[%s:%s]", "" /* eb.env.Region */, eb.name)
}

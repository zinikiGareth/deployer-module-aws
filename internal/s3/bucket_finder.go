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

type bucketFinder struct {
	tools *pluggable.Tools

	loc  *errorsink.Location
	name string

	client        *s3.Client
	alreadyExists bool
	// cloud *BucketCloud
}

func (b *bucketFinder) Loc() *errorsink.Location {
	return b.loc
}

func (b *bucketFinder) ShortDescription() string {
	return "aws.s3.Bucket[" + b.name + "]"
}

func (b *bucketFinder) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.s3.Bucket[")
	iw.AttrsWhere(b)
	iw.TextAttr("named", b.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (b *bucketFinder) BuildModel(pres pluggable.ValuePresenter) {
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

	// TODO: I think we should capture the cloud bucket
	// (but maybe not, maybe the name is all we need)
	pres.Present(b)
}

func (b *bucketFinder) ObtainDest() files.FileDest {
	return NewBucketTransfer(b.client, b.name)
}

func (b *bucketFinder) String() string {
	return fmt.Sprintf("FindBucket[%s:%s]", "" /* eb.env.Region */, b.name)
}

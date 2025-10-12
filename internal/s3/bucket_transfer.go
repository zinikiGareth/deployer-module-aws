package s3

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"ziniki.org/deployer/coremod/pkg/corebottom"
)

type bucketTransfer struct {
	client *s3.Client
	bucket string
	path   string
}

func (b *bucketTransfer) PourInto(key string, contents io.Reader) error {
	log.Printf("want to pour %s into %s:%s", key, b.bucket, b.path+key)
	_, err := b.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.path + key),
		Body:   contents,
	})
	return err
}

func (b *bucketTransfer) Relative(name string) (corebottom.FileDest, error) {
	nested := &bucketTransfer{client: b.client, bucket: b.bucket, path: b.path + name + "/"}
	return nested, nil
}

func NewBucketTransfer(client *s3.Client, bucket string) *bucketTransfer {
	return &bucketTransfer{client: client, bucket: bucket}
}

var _ corebottom.FileDest = &bucketTransfer{}

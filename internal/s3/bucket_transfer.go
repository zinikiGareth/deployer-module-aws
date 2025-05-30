package s3

import (
	"context"
	"io"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type bucketTransfer struct {
	client *s3.Client
	bucket string
}

func (b *bucketTransfer) PourInto(key string, contents io.Reader) {
	log.Printf("want to pour %s into %s", key, b.bucket)
	b.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
		Body:   contents,
	})
}

func NewBucketTransfer(client *s3.Client, bucket string) *bucketTransfer {
	return &bucketTransfer{client: client, bucket: bucket}
}

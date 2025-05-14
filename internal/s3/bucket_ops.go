package s3

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func CreateBucket(client *s3.Client, name string) *s3.CreateBucketOutput {
	bucket, err := client.CreateBucket(context.TODO(), &s3.CreateBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		log.Fatalf("error creating bucket %s: %v\n", name, err)
		return nil
	} else {
		err = s3.NewBucketExistsWaiter(client).Wait(
			context.TODO(), &s3.HeadBucketInput{Bucket: aws.String(name)},
			time.Minute,
		)
		if err != nil {
			log.Printf("Failed attempt to wait for bucket %s to exist: %v.\n", name, err)
		}
	}
	return bucket

}

func EmptyBucket() {

}

func DeleteBucket(client *s3.Client, name string) {
	out, err := client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		log.Fatalf("error deleting bucket %s: %v\n", name, err)
	} else {
		err = s3.NewBucketNotExistsWaiter(client).Wait(
			context.TODO(), &s3.HeadBucketInput{Bucket: aws.String(name)},
			time.Minute,
		)
		if err != nil {
			log.Printf("Failed attempt to wait for bucket %s to be deleted: %v.\n", name, err)
		}
		log.Printf("Deleted bucket %s: %v\n", name, out)
	}

}

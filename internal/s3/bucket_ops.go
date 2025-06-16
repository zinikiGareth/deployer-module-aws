package s3

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
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

func EmptyBucket(client *s3.Client, name string) bool {
	for {
		keys, err := ListBucket(client, name)
		if err != nil {
			log.Printf("failed to list bucket %s", name)
			return false
		}
		if keys != nil {
			DeleteFromBucket(client, name, keys)
		} else {
			return true
		}
	}
}

func ListBucket(client *s3.Client, name string) ([]types.ObjectIdentifier, error) {
	objs, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(name),
	})
	if err != nil {
		return nil, err
	}

	if *objs.KeyCount == 0 {
		return nil, nil
	}
	return identifiersOf(objs.Contents), nil
}

func DeleteFromBucket(client *s3.Client, name string, keys []types.ObjectIdentifier) error {
	_, err := client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(name),
		Delete: &types.Delete{
			Objects: keys,
			Quiet:   aws.Bool(true),
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func DeleteBucket(client *s3.Client, name string) bool {
	_, err := client.DeleteBucket(context.TODO(), &s3.DeleteBucketInput{
		Bucket: aws.String(name),
	})
	if err != nil {
		log.Fatalf("error deleting bucket %s: %v\n", name, err)
		return false
	} else {
		err = s3.NewBucketNotExistsWaiter(client).Wait(
			context.TODO(), &s3.HeadBucketInput{Bucket: aws.String(name)},
			time.Minute,
		)
		if err != nil {
			log.Printf("Failed attempt to wait for bucket %s to be deleted: %v.\n", name, err)
		}
		log.Printf("Deleted bucket %s\n", name)
	}
	return true
}

func identifiersOf(objs []types.Object) []types.ObjectIdentifier {
	ret := make([]types.ObjectIdentifier, len(objs))
	for k, o := range objs {
		ret[k] = types.ObjectIdentifier{Key: o.Key}
	}
	return ret
}

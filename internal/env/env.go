package env

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AwsEnv struct {
	cfg      aws.Config
	s3client *s3.Client
}

func (a *AwsEnv) Init() {
	var err error
	a.cfg, err = config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	a.s3client = s3.NewFromConfig(a.cfg)

}

func (a *AwsEnv) S3Client() *s3.Client {
	return a.s3client
}

func InitAwsEnv() *AwsEnv {
	ret := &AwsEnv{}
	ret.Init()
	return ret
}

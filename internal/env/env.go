package env

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/acm"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53domains"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type AwsEnv struct {
	cfg                  aws.Config
	acmclient            *acm.Client
	cfclient             *cloudfront.Client
	dynamoClient         *dynamodb.Client
	iamclient            *iam.Client
	neptuneClient        *neptune.Client
	route53client        *route53.Client
	route53domainsclient *route53domains.Client
	s3client             *s3.Client
}

func (a *AwsEnv) Init() {
	var err error
	a.cfg, err = config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	a.acmclient = acm.NewFromConfig(a.cfg)
	a.cfclient = cloudfront.NewFromConfig(a.cfg)
	a.dynamoClient = dynamodb.NewFromConfig(a.cfg)
	a.iamclient = iam.NewFromConfig(a.cfg)
	a.neptuneClient = neptune.NewFromConfig(a.cfg)
	a.route53client = route53.NewFromConfig(a.cfg)
	a.route53domainsclient = route53domains.NewFromConfig(a.cfg)
	a.s3client = s3.NewFromConfig(a.cfg)
}

func (a *AwsEnv) ACMClient() *acm.Client {
	return a.acmclient
}

func (a *AwsEnv) CFClient() *cloudfront.Client {
	return a.cfclient
}

func (a *AwsEnv) DynamoClient() *dynamodb.Client {
	return a.dynamoClient
}

func (a *AwsEnv) IAMClient() *iam.Client {
	return a.iamclient
}

func (a *AwsEnv) NeptuneClient() *neptune.Client {
	return a.neptuneClient
}

func (a *AwsEnv) Route53Client() *route53.Client {
	return a.route53client
}

func (a *AwsEnv) Route53DomainsClient() *route53domains.Client {
	return a.route53domainsclient
}

func (a *AwsEnv) S3Client() *s3.Client {
	return a.s3client
}

func InitAwsEnv() *AwsEnv {
	ret := &AwsEnv{}
	ret.Init()
	return ret
}

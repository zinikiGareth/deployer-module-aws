package awsmod

import (
	"reflect"

	"ziniki.org/deployer/deployer/pkg/deployer"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/acm"
	"ziniki.org/deployer/modules/aws/internal/env"
	"ziniki.org/deployer/modules/aws/internal/route53"
	"ziniki.org/deployer/modules/aws/internal/s3"
)

// var testRunner deployer.TestRunner

func ProvideTestRunner(runner deployer.TestRunner) error {
	// testRunner = runner
	return nil
}

func RegisterWithDeployer(deployer deployer.Deployer) error {
	tools := deployer.ObtainTools()
	tools.Register.ProvideDriver("aws.AwsEnv", env.InitAwsEnv())

	tools.Register.Register(reflect.TypeFor[pluggable.Blank](), "aws.Route53.DomainName", &route53.DomainNameBlank{})
	tools.Register.Register(reflect.TypeFor[pluggable.Blank](), "aws.CertificateManager.Certificate", &acm.CertificateBlank{})
	tools.Register.Register(reflect.TypeFor[pluggable.Blank](), "aws.S3.Bucket", &s3.BucketBlank{})

	return nil
}

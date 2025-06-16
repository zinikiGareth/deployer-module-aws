package awsmod

import (
	"ziniki.org/deployer/deployer/pkg/creator"
	"ziniki.org/deployer/deployer/pkg/deployer"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/acm"
	"ziniki.org/deployer/modules/aws/internal/cfront"
	"ziniki.org/deployer/modules/aws/internal/env"
	"ziniki.org/deployer/modules/aws/internal/iam"
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

	tools.Register.Register("blank", "aws.Route53.DomainName", &route53.DomainNameBlank{})
	tools.Register.Register("blank", "aws.Route53.CNAME", &route53.CNAMEBlank{})
	tools.Register.Register("blank", "aws.CertificateManager.Certificate", &acm.CertificateBlank{})
	tools.Register.Register("blank", "aws.CloudFront.Distribution", &cfront.DistributionBlank{})
	tools.Register.Register("blank", "aws.S3.Bucket", &s3.BucketBlank{})
	tools.Register.Register("blank", "aws.IAM.Policy", &iam.PolicyBlank{})

	// Permissions by name
	tools.Repository.IntroduceSymbol(pluggable.SymbolName("aws.action.S3.GetObject"), creator.MakeString("s3:GetObject"))
	tools.Repository.IntroduceSymbol(pluggable.SymbolName("aws.action.S3.PutObject"), creator.MakeString("s3:PutObject"))

	// Principals
	tools.Repository.IntroduceSymbol(pluggable.SymbolName("aws.principal.AWS"), creator.MakeString("AWS"))
	tools.Repository.IntroduceSymbol(pluggable.SymbolName("aws.principal.Service"), creator.MakeString("Service"))

	// Service Principals
	tools.Repository.IntroduceSymbol(pluggable.SymbolName("aws.principal.CloudFront"), creator.MakeString("cloudfront.amazonaws.com"))

	// other strings
	tools.Repository.IntroduceSymbol(pluggable.SymbolName("aws.cond.StringEquals"), creator.MakeString("StringEquals"))
	tools.Repository.IntroduceSymbol(pluggable.SymbolName("aws.SourceArn"), creator.MakeString("aws:SourceArn"))

	return nil
}

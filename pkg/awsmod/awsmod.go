package awsmod

import (
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/modules/aws/internal/acm"
	"ziniki.org/deployer/modules/aws/internal/cfront"
	"ziniki.org/deployer/modules/aws/internal/env"
	"ziniki.org/deployer/modules/aws/internal/iam"
	"ziniki.org/deployer/modules/aws/internal/route53"
	"ziniki.org/deployer/modules/aws/internal/s3"
)

// var testRunner deployer.TestRunner

func ProvideTestRunner(runner driverbottom.TestRunner) error {
	// testRunner = runner
	return nil
}

func RegisterWithDriver(deployer driverbottom.Driver) error {
	tools := deployer.ObtainCoreTools()
	tools.Register.ProvideDriver("aws.AwsEnv", env.InitAwsEnv())

	tools.Register.Register("blank", "aws.Route53.DomainName", &route53.DomainNameBlank{})
	tools.Register.Register("blank", "aws.Route53.CNAME", &route53.CNAMEBlank{})
	tools.Register.Register("blank", "aws.CertificateManager.Certificate", &acm.CertificateBlank{})
	tools.Register.Register("blank", "aws.CloudFront.OriginAccessControl", &cfront.OACBlank{})
	tools.Register.Register("blank", "aws.CloudFront.ResponseHeadersPolicy", &cfront.RHPBlank{})
	tools.Register.Register("blank", "aws.CloudFront.CacheBehavior", &cfront.CacheBehaviorBlank{})
	tools.Register.Register("blank", "aws.CloudFront.CachePolicy", &cfront.CachePolicyBlank{})
	tools.Register.Register("blank", "aws.CloudFront.Distribution", &cfront.DistributionBlank{})
	tools.Register.Register("blank", "aws.S3.Bucket", &s3.BucketBlank{})
	tools.Register.Register("blank", "aws.IAM.Policy", &iam.PolicyBlank{})

	// Permissions by name
	tools.Repository.IntroduceSymbol(driverbottom.SymbolName("aws.action.S3.GetObject"), drivertop.MakeString("s3:GetObject"))
	tools.Repository.IntroduceSymbol(driverbottom.SymbolName("aws.action.S3.PutObject"), drivertop.MakeString("s3:PutObject"))

	// Principals
	tools.Repository.IntroduceSymbol(driverbottom.SymbolName("aws.principal.AWS"), drivertop.MakeString("AWS"))
	tools.Repository.IntroduceSymbol(driverbottom.SymbolName("aws.principal.Service"), drivertop.MakeString("Service"))

	// Service Principals
	tools.Repository.IntroduceSymbol(driverbottom.SymbolName("aws.principal.CloudFront"), drivertop.MakeString("cloudfront.amazonaws.com"))

	// other strings
	tools.Repository.IntroduceSymbol(driverbottom.SymbolName("aws.cond.StringEquals"), drivertop.MakeString("StringEquals"))
	tools.Repository.IntroduceSymbol(driverbottom.SymbolName("aws.SourceArn"), drivertop.MakeString("aws:SourceArn"))

	return nil
}

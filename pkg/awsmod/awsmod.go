package awsmod

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/acm"
	"ziniki.org/deployer/modules/aws/internal/cfront"
	"ziniki.org/deployer/modules/aws/internal/dynamodb"
	"ziniki.org/deployer/modules/aws/internal/env"
	"ziniki.org/deployer/modules/aws/internal/iam"
	"ziniki.org/deployer/modules/aws/internal/neptune"
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

	mytools := tools.RetrieveOther("coremod").(*corebottom.Tools)

	tools.Register.Register("target", "cloudfront.distribution.fromS3", cfront.NewWebsiteFromS3Handler(mytools))

	tools.Register.Register("blank", "aws.Route53.DomainName", &route53.DomainNameBlank{})
	tools.Register.Register("blank", "aws.Route53.ALIAS", &route53.ALIASBlank{})
	tools.Register.Register("blank", "aws.Route53.CNAME", &route53.CNAMEBlank{})
	tools.Register.Register("blank", "aws.CertificateManager.Certificate", &acm.CertificateBlank{})
	tools.Register.Register("blank", "aws.CloudFront.OriginAccessControl", &cfront.OACBlank{})
	tools.Register.Register("blank", "aws.CloudFront.ResponseHeadersPolicy", &cfront.RHPBlank{})
	tools.Register.Register("blank", "aws.CloudFront.CacheBehavior", &cfront.CacheBehaviorBlank{})
	tools.Register.Register("blank", "aws.CloudFront.CachePolicy", &cfront.CachePolicyBlank{})
	tools.Register.Register("blank", "aws.CloudFront.Distribution", &cfront.DistributionBlank{})
	tools.Register.Register("blank", "aws.S3.Bucket", &s3.BucketBlank{})
	tools.Register.Register("blank", "aws.IAM.Policy", &iam.PolicyBlank{})
	tools.Register.Register("blank", "aws.Neptune.SubnetGroup", &neptune.SubnetBlank{})
	tools.Register.Register("blank", "aws.Neptune.Cluster", &neptune.ClusterBlank{})
	tools.Register.Register("blank", "aws.DynamoDB.Table", &dynamodb.TableBlank{})

	loc := &errorsink.Location{}
	// Permissions by name
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.action.S3.GetObject"), drivertop.MakeString(loc, "s3:GetObject"))
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.action.S3.PutObject"), drivertop.MakeString(loc, "s3:PutObject"))

	// Principals
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.principal.AWS"), drivertop.MakeString(loc, "AWS"))
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.principal.Service"), drivertop.MakeString(loc, "Service"))

	// Service Principals
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.principal.CloudFront"), drivertop.MakeString(loc, "cloudfront.amazonaws.com"))

	// other strings
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.cond.StringEquals"), drivertop.MakeString(loc, "StringEquals"))
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.SourceArn"), drivertop.MakeString(loc, "aws:SourceArn"))

	return nil
}

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
	"ziniki.org/deployer/modules/aws/internal/gatewayV2"
	"ziniki.org/deployer/modules/aws/internal/iam"
	"ziniki.org/deployer/modules/aws/internal/lambda"
	"ziniki.org/deployer/modules/aws/internal/neptune"
	"ziniki.org/deployer/modules/aws/internal/route53"
	"ziniki.org/deployer/modules/aws/internal/s3"
	"ziniki.org/deployer/modules/aws/internal/vpc"
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

	tools.Register.ExtensionPoint("dns-asserter")

	tools.Register.Register("target", "cloudfront.distribution.fromS3", cfront.NewWebsiteFromS3Handler(mytools))
	tools.Register.Register("target", "cloudfront.invalidate", cfront.NewInvalidateHandler(mytools))

	tools.Register.Register("target", "lambda.function", lambda.NewLambdaFunction(mytools))
	tools.Register.Register("target", "lambda.addPermissions", lambda.AddLambdaPermissions(mytools))

	tools.Register.Register("target", "api.gatewayV2", gatewayV2.NewAPI(mytools))

	tools.Register.Register("blank", "aws.ApiGatewayV2.Api", &gatewayV2.ApiBlank{})
	tools.Register.Register("blank", "aws.ApiGatewayV2.Deployment", &gatewayV2.DeploymentBlank{})
	tools.Register.Register("blank", "aws.ApiGatewayV2.Integration", &gatewayV2.IntegrationBlank{})
	tools.Register.Register("blank", "aws.ApiGatewayV2.Route", &gatewayV2.RouteBlank{})
	tools.Register.Register("blank", "aws.ApiGatewayV2.Stage", &gatewayV2.StageBlank{})
	tools.Register.Register("blank", "aws.ApiGatewayV2.VPCLink", &gatewayV2.VPCLinkBlank{})
	tools.Register.Register("blank", "aws.CertificateManager.Certificate", &acm.CertificateBlank{})
	tools.Register.Register("blank", "aws.CloudFront.OriginAccessControl", &cfront.OACBlank{})
	tools.Register.Register("blank", "aws.CloudFront.ResponseHeadersPolicy", &cfront.RHPBlank{})
	tools.Register.Register("blank", "aws.CloudFront.CacheBehavior", &cfront.CacheBehaviorBlank{})
	tools.Register.Register("blank", "aws.CloudFront.CachePolicy", &cfront.CachePolicyBlank{})
	tools.Register.Register("blank", "aws.CloudFront.Distribution", &cfront.DistributionBlank{})
	tools.Register.Register("blank", "aws.DynamoDB.Table", &dynamodb.TableBlank{})
	tools.Register.Register("blank", "aws.IAM.Policy", &iam.PolicyBlank{})
	tools.Register.Register("blank", "aws.IAM.Role", &iam.RoleBlank{})
	tools.Register.Register("blank", "aws.Lambda.Function", &lambda.FunctionBlank{})
	tools.Register.Register("blank", "aws.Neptune.SubnetGroup", &neptune.SubnetBlank{})
	tools.Register.Register("blank", "aws.Neptune.Cluster", &neptune.ClusterBlank{})
	tools.Register.Register("blank", "aws.Neptune.Instance", &neptune.InstanceBlank{})
	tools.Register.Register("blank", "aws.Route53.DomainName", &route53.DomainNameBlank{})
	tools.Register.Register("blank", "aws.Route53.ALIAS", &route53.ALIASBlank{})
	tools.Register.Register("blank", "aws.Route53.CNAME", &route53.CNAMEBlank{})
	tools.Register.Register("blank", "aws.S3.Bucket", &s3.BucketBlank{})
	tools.Register.Register("blank", "aws.VPC.VPC", &vpc.VPCBlank{})

	tools.Register.Register("prop-interpreter", "aws.DynamoFields", driverbottom.CreateInterpreter(dynamodb.CreateFieldInterpreter))
	tools.Register.Register("prop-interpreter", "aws.IAM.WithRole", driverbottom.CreateInterpreter(iam.CreateWithRoleInterpreter))
	tools.Register.Register("prop-interpreter", "aws.S3.Location", driverbottom.CreateInterpreter(s3.CreateLocationInterpreter))
	tools.Register.Register("prop-interpreter", "aws.VPC.Config", driverbottom.CreateInterpreter(vpc.CreateConfigInterpreter))

	loc := &errorsink.Location{}
	// Permissions by name
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.action.APIGateway.GET"), drivertop.MakeString(loc, "apigateway:GET"))
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.action.ec2.CreateNetworkInterface"), drivertop.MakeString(loc, "ec2:CreateNetworkInterface"))
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.action.ec2.DescribeNetworkInterfaces"), drivertop.MakeString(loc, "ec2:DescribeNetworkInterfaces"))
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.action.ec2.DeleteNetworkInterface"), drivertop.MakeString(loc, "ec2:DeleteNetworkInterface"))
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.action.S3.GetObject"), drivertop.MakeString(loc, "s3:GetObject"))
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.action.S3.PutObject"), drivertop.MakeString(loc, "s3:PutObject"))

	// Principals
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.principal.AWS"), drivertop.MakeString(loc, "AWS"))
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.principal.Service"), drivertop.MakeString(loc, "Service"))

	// Generic Resource ARN patterns
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.resource.APIGatewayV2"), drivertop.MakeString(loc, "arn:aws:apigateway:us-east-1::/apis"))

	// Service Principals
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.principal.CloudFront"), drivertop.MakeString(loc, "cloudfront.amazonaws.com"))

	// other strings
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.cond.StringEquals"), drivertop.MakeString(loc, "StringEquals"))
	tools.Repository.TopScope().IntroduceSymbol(driverbottom.SymbolName("aws.SourceArn"), drivertop.MakeString(loc, "aws:SourceArn"))

	return nil
}

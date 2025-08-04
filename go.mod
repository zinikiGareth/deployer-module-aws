module ziniki.org/deployer/modules/aws

go 1.24.5

require (
	github.com/aws/aws-sdk-go-v2 v1.37.1
	github.com/aws/aws-sdk-go-v2/config v1.29.14
	github.com/aws/aws-sdk-go-v2/service/acm v1.32.0
	github.com/aws/aws-sdk-go-v2/service/cloudfront v1.46.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.44.0
	github.com/aws/aws-sdk-go-v2/service/iam v1.42.1
	github.com/aws/aws-sdk-go-v2/service/neptune v1.37.3
	github.com/aws/aws-sdk-go-v2/service/route53 v1.51.1
	github.com/aws/aws-sdk-go-v2/service/route53domains v1.29.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.79.3
	github.com/aws/smithy-go v1.22.5
	ziniki.org/deployer/coremod v0.0.0
	ziniki.org/deployer/driver v0.0.0
)

require (
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.6.11 // indirect
	github.com/aws/aws-sdk-go-v2/credentials v1.17.67 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.16.30 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.4.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.8.3 // indirect
	github.com/aws/aws-sdk-go-v2/internal/v4a v1.3.34 // indirect
	github.com/aws/aws-sdk-go-v2/service/apigatewayv2 v1.29.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.239.0
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/checksum v1.7.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/endpoint-discovery v1.10.17 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.13.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.18.15 //indirect
	github.com/aws/aws-sdk-go-v2/service/lambda v1.73.0
	github.com/aws/aws-sdk-go-v2/service/sso v1.25.3 // indirect
	github.com/aws/aws-sdk-go-v2/service/ssooidc v1.30.1 // indirect
	github.com/aws/aws-sdk-go-v2/service/sts v1.33.19
)

replace ziniki.org/deployer/driver => ../deployer/driver

replace ziniki.org/deployer/coremod => ../deployer/coremod

package cfront

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type distributionCreator struct {
	tools *pluggable.Tools

	loc      *errorsink.Location
	name     string
	teardown pluggable.TearDown

	client        *cloudfront.Client
	oacId         string
	cpId          string
	alreadyExists bool
	arn           string
	props         map[pluggable.Identifier]pluggable.Expr
}

func (cfdc *distributionCreator) Loc() *errorsink.Location {
	return cfdc.loc
}

func (cfdc *distributionCreator) ShortDescription() string {
	return "aws.CloudFront.Distribution[" + cfdc.name + "]"
}

func (cfdc *distributionCreator) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.CloudFront.Distribution[")
	iw.AttrsWhere(cfdc)
	iw.TextAttr("named", cfdc.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (cfdc *distributionCreator) BuildModel(pres pluggable.ValuePresenter) {
	/*
		domainExpr := find(cfdc.props, "Domain")
		if domainExpr == nil {
			log.Fatalf("must specify a domain instance to create a certificate")
		}
		domainObj := domainExpr.Eval(cfdc.tools.Storage)
		log.Printf("%T %p %v\n", domainObj, domainObj, domainObj)
		domain, ok := domainObj.(myroute53.ExportedDomain)
		if !ok {
			log.Fatalf("Domain did not point to a domain instance")
		}
		cfdc.hzid = domain.HostedZoneId()
		if cfdc.validationMethod == "" {
			cfdc.validationMethod = types.ValidationMethodDns
		}
	*/
	eq := cfdc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	cfdc.client = awsEnv.CFClient()

	distros, err := cfdc.client.ListDistributions(context.TODO(), &cloudfront.ListDistributionsInput{})
	if err != nil {
		log.Fatalf("could not list OACs")
	}
	for _, p := range distros.DistributionList.Items {
		tags, err := cfdc.client.ListTagsForResource(context.TODO(), &cloudfront.ListTagsForResourceInput{Resource: p.ARN})
		if err != nil {
			log.Fatalf("error trying to obtain tags for %s\n", *p.ARN)
		}
		for _, q := range tags.Tags.Items {
			if q.Key != nil && *q.Key == "deployer-name" && q.Value != nil && *q.Value == cfdc.name {
				cfdc.arn = *p.ARN
				cfdc.alreadyExists = true
			}
		}
	}

	if !cfdc.alreadyExists {
		oaccname := "oac-name"
		fred, err := cfdc.client.ListOriginAccessControls(context.TODO(), &cloudfront.ListOriginAccessControlsInput{})
		if err != nil {
			log.Fatalf("could not list OACs")
		}
		for _, p := range fred.OriginAccessControlList.Items {
			if p.Id != nil && p.Name != nil && *p.Name == oaccname {
				cfdc.oacId = *p.Id
				log.Printf("found OAC for %s with id %s\n", oaccname, cfdc.oacId)
			}
		}

		cpname := "cp-name"
		bert, err := cfdc.client.ListCachePolicies(context.TODO(), &cloudfront.ListCachePoliciesInput{})
		if err != nil {
			log.Fatalf("could not list CPs")
		}
		for _, p := range bert.CachePolicyList.Items {
			if p.CachePolicy.Id != nil && p.CachePolicy.CachePolicyConfig.Name != nil && *p.CachePolicy.CachePolicyConfig.Name == cpname {
				cfdc.cpId = *p.CachePolicy.Id
				log.Printf("found CachePolicy for %s with id %s\n", cpname, cfdc.cpId)
			}
		}
	}

	/*
		cfdc.route53 = awsEnv.Route53Client()

		certs := cfdc.findCertificatesFor(cfdc.name)
		if len(certs) == 0 {
			log.Printf("there were no certs found for %s\n", cfdc.name)
		} else {
			log.Printf("found %d certs for %s\n", len(certs), cfdc.name)
			cfdc.alreadyExists = true
			cfdc.arn = certs[0]
			cfdc.describeCertificate(cfdc.arn)
		}

	*/
	// TODO: do we need to capture something here?
	pres.Present(cfdc)
}

func (cfdc *distributionCreator) UpdateReality() {
	if cfdc.alreadyExists {
		log.Printf("distribution %s already existed for %s\n", cfdc.arn, cfdc.name)
		return
	}
	comment := "we should probably have this be a required parameter"

	if cfdc.oacId == "" {
		oaccname := "oac-name"
		oacComment := "OAC for " + comment
		oaccfg := types.OriginAccessControlConfig{Name: &oaccname, OriginAccessControlOriginType: types.OriginAccessControlOriginTypesS3, SigningBehavior: types.OriginAccessControlSigningBehaviorsAlways, SigningProtocol: types.OriginAccessControlSigningProtocolsSigv4, Description: &oacComment}
		oac, err := cfdc.client.CreateOriginAccessControl(context.TODO(), &cloudfront.CreateOriginAccessControlInput{OriginAccessControlConfig: &oaccfg})
		if err != nil {
			log.Fatalf("failed to create OAC for %s: %v\n", cfdc.name, err)
		}
		cfdc.oacId = *oac.OriginAccessControl.Id
	}

	if cfdc.cpId == "" {
		cpname := "cp-name"
		cpComment := "CP for " + comment
		var minttl int64 = 300
		cpc := types.CachePolicyConfig{Name: &cpname, Comment: &cpComment, MinTTL: &minttl}
		oac, err := cfdc.client.CreateCachePolicy(context.TODO(), &cloudfront.CreateCachePolicyInput{CachePolicyConfig: &cpc})
		if err != nil {
			log.Fatalf("failed to create CachePolicy for %s: %v\n", cfdc.name, err)
		}
		cfdc.cpId = *oac.CachePolicy.Id
	}

	origindns := "news.consolidator.info.s3.us-east-1.amazonaws.com"
	fred := "a-unique-id" + cfdc.name
	dcb := types.DefaultCacheBehavior{TargetOriginId: &fred, ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps, CachePolicyId: &cfdc.cpId}
	e := true
	// dn := "www.consolidator.news"
	empty := ""
	s3orig := types.S3OriginConfig{OriginAccessIdentity: &empty}
	items := []types.Origin{{DomainName: &origindns, Id: &fred, OriginAccessControlId: &cfdc.oacId, S3OriginConfig: &s3orig}}
	quant := int32(len(items))
	config := types.DistributionConfig{CallerReference: &cfdc.name, Comment: &comment, DefaultCacheBehavior: &dcb, Enabled: &e, Origins: &types.Origins{Items: items, Quantity: &quant}}
	tagkey := "deployer-name"
	tags := types.Tags{Items: []types.Tag{{Key: &tagkey, Value: &cfdc.name}}}
	req, err := cfdc.client.CreateDistributionWithTags(context.TODO(), &cloudfront.CreateDistributionWithTagsInput{DistributionConfigWithTags: &types.DistributionConfigWithTags{DistributionConfig: &config, Tags: &tags}})
	if err != nil {
		log.Fatalf("failed to create distribution %s: %v\n", cfdc.name, err)
	}
	log.Printf("created distribution %s: %s\n", cfdc.name, *req.Distribution.ARN)
	cfdc.arn = *req.Distribution.ARN
	/*
		cfdc.describeCertificate(*req.CertificateArn)
		var waitFor time.Duration = 1
		for {
			log.Printf("sleeping for %ds\n", waitFor)
			time.Sleep(waitFor * time.Second)
			if cfdc.tryToValidateCert(*req.CertificateArn) {
				break
			}
			waitFor = min(2*waitFor, 60)
			fmt.Printf("still pending validation; wait another %ds\n", waitFor)
		}
	*/
}

func (cfdc *distributionCreator) TearDown() {
	/*
		if !cfdc.alreadyExists {
			log.Printf("no certificate existed for %s\n", cfdc.name)
			return
		}
		log.Printf("you have asked to tear down certificate for %s (arn: %s) with mode %s\n", cfdc.name, cfdc.arn, cfdc.teardown.Mode())
		switch cfdc.teardown.Mode() {
		case "preserve":
			log.Printf("not deleting certificate %s because teardown mode is 'preserve'", cfdc.name)
		case "delete":
			log.Printf("deleting certificate for %s with teardown mode 'delete'", cfdc.name)
			DeleteCertificate(cfdc.client, cfdc.arn)
		default:
			log.Printf("cannot handle teardown mode '%s' for bucket %s", cfdc.teardown.Mode(), cfdc.name)
		}
	*/
}

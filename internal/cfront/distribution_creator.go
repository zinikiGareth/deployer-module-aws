package cfront

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type distributionCreator struct {
	tools *pluggable.Tools

	loc        *errorsink.Location
	name       string
	domain     pluggable.Expr
	viewerCert pluggable.Expr
	teardown   pluggable.TearDown

	client        *cloudfront.Client
	oacId         string
	cpId          string
	rpId          string
	distroId      string
	alreadyExists bool
	arn           string
	domainName    string
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

func (cfdc *distributionCreator) BuildModel(pres pluggable.ValuePresenter) {
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
				cfdc.distroId = *p.Id
				cfdc.domainName = *p.DomainName
				cfdc.alreadyExists = true
				log.Printf("found distro %s: %s %s %s\n", cfdc.name, cfdc.arn, cfdc.distroId, cfdc.domainName)
			}
		}
	}

	if !cfdc.alreadyExists || cfdc.tools.Options.TearDown {
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

		rhname := "fred"
		zeb, err := cfdc.client.ListResponseHeadersPolicies(context.TODO(), &cloudfront.ListResponseHeadersPoliciesInput{})
		if err != nil {
			log.Fatalf("could not list RHPs")
		}
		for _, p := range zeb.ResponseHeadersPolicyList.Items {
			if p.ResponseHeadersPolicy.Id != nil {
				rhc, err := cfdc.client.GetResponseHeadersPolicyConfig(context.TODO(), &cloudfront.GetResponseHeadersPolicyConfigInput{Id: p.ResponseHeadersPolicy.Id})
				if err != nil {
					log.Fatalf("could not recover RHP %s", *p.ResponseHeadersPolicy.Id)
				}
				if rhc.ResponseHeadersPolicyConfig.Name != nil && *rhc.ResponseHeadersPolicyConfig.Name == rhname {
					cfdc.rpId = *p.ResponseHeadersPolicy.Id
					log.Printf("found rhpc %s\n", cfdc.rpId)
				}
			}
		}
	}

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

	/* More stuff hacked in that shouldn't be ... */
	// Again, probably wants to be found somewhere and passed in ...
	rpname := "fred"
	ct := "Content-Type"
	ov := true
	th := "text/html"
	rhs := []types.ResponseHeadersPolicyCustomHeader{{Header: &ct, Override: &ov, Value: &th}}
	rhslen := int32(len(rhs))
	ch := types.ResponseHeadersPolicyCustomHeadersConfig{Items: rhs, Quantity: &rhslen}
	rhp := types.ResponseHeadersPolicyConfig{Name: &rpname, CustomHeadersConfig: &ch}
	crhp, err := cfdc.client.CreateResponseHeadersPolicy(context.TODO(), &cloudfront.CreateResponseHeadersPolicyInput{ResponseHeadersPolicyConfig: &rhp})
	if err != nil {
		log.Fatalf("failed to create CRHP for %s (%s): %v\n", cfdc.name, rpname, err)
	}
	rpid := *crhp.ResponseHeadersPolicy.Id

	origindns := "news.consolidator.info.s3.us-east-1.amazonaws.com"
	targetOriginId := "a-unique-id" + cfdc.name
	dcb := types.DefaultCacheBehavior{TargetOriginId: &targetOriginId, ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps, CachePolicyId: &cfdc.cpId}
	e := true
	empty := ""
	s3orig := types.S3OriginConfig{OriginAccessIdentity: &empty}
	origins := []types.Origin{{DomainName: &origindns, Id: &targetOriginId, OriginAccessControlId: &cfdc.oacId, S3OriginConfig: &s3orig}}
	nOrigins := int32(len(origins))

	forHtml := "*.html"
	htmlBehavior := types.CacheBehavior{TargetOriginId: &targetOriginId, PathPattern: &forHtml, ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps, CachePolicyId: &cfdc.cpId, ResponseHeadersPolicyId: &rpid}
	cbs := []types.CacheBehavior{htmlBehavior}
	cbl := int32(len(cbs))

	behaviors := types.CacheBehaviors{Quantity: &cbl, Items: cbs}
	domain := cfdc.tools.Storage.EvalAsString(cfdc.domain)

	aliases := []string{domain}
	nAliases := int32(len(aliases))
	rootObject := "index.html"
	config := types.DistributionConfig{CallerReference: &cfdc.name, Comment: &comment, DefaultCacheBehavior: &dcb, CacheBehaviors: &behaviors, Enabled: &e, DefaultRootObject: &rootObject, Origins: &types.Origins{Items: origins, Quantity: &nOrigins}, Aliases: &types.Aliases{Items: aliases, Quantity: &nAliases}}
	if cfdc.viewerCert != nil {
		vc := cfdc.tools.Storage.Eval(cfdc.viewerCert)
		log.Printf("have vc %T %v\n", vc, vc)
		vcs, ok := vc.(string)
		if !ok {
			tmp, ok := vc.(fmt.Stringer)
			if !ok {
				log.Fatalf("not a string or Stringer but %T", vc)
			}
			vcs = tmp.String()
		}
		log.Printf("have cert arn %s\n", vcs)
		minver := types.MinimumProtocolVersionTLSv122021
		supp := types.SSLSupportMethodSniOnly
		cfdef := false
		config.ViewerCertificate = &types.ViewerCertificate{ACMCertificateArn: &vcs, MinimumProtocolVersion: minver, SSLSupportMethod: supp, CloudFrontDefaultCertificate: &cfdef}
	}
	tagkey := "deployer-name"
	tags := types.Tags{Items: []types.Tag{{Key: &tagkey, Value: &cfdc.name}}}
	req, err := cfdc.client.CreateDistributionWithTags(context.TODO(), &cloudfront.CreateDistributionWithTagsInput{DistributionConfigWithTags: &types.DistributionConfigWithTags{DistributionConfig: &config, Tags: &tags}})
	if err != nil {
		log.Fatalf("failed to create distribution %s: %v\n", cfdc.name, err)
	}
	log.Printf("created distribution %s: %s %s %s\n", cfdc.name, *req.Distribution.ARN, *req.Distribution.Id, *req.Distribution.DomainName)
	cfdc.arn = *req.Distribution.ARN
	cfdc.distroId = *req.Distribution.Id
	cfdc.domainName = *req.Distribution.DomainName
}

func (cfdc *distributionCreator) TearDown() {
	if cfdc.alreadyExists {
		log.Printf("you have asked to tear down distribution %s (id: %s, arn: %s) with mode %s\n", cfdc.name, cfdc.distroId, cfdc.arn, cfdc.teardown.Mode())

		cfdc.DisableIt()
		cfdc.DeleteIt()
	} else {
		log.Printf("no distribution existed for %s\n", cfdc.name)
	}

	if cfdc.cpId != "" {
		x, err := cfdc.client.GetCachePolicy(context.TODO(), &cloudfront.GetCachePolicyInput{Id: &cfdc.cpId})
		if err != nil {
			log.Fatalf("could not get CP %s: %v", cfdc.cpId, err)
		}
		_, err = cfdc.client.DeleteCachePolicy(context.TODO(), &cloudfront.DeleteCachePolicyInput{Id: &cfdc.cpId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete CP %s: %v", cfdc.cpId, err)
		}
		log.Printf("deleted CP %s\n", cfdc.cpId)
	}

	if cfdc.rpId != "" {
		x, err := cfdc.client.GetResponseHeadersPolicy(context.TODO(), &cloudfront.GetResponseHeadersPolicyInput{Id: &cfdc.rpId})
		if err != nil {
			log.Fatalf("could not get RHP %s: %v", cfdc.rpId, err)
		}
		_, err = cfdc.client.DeleteResponseHeadersPolicy(context.TODO(), &cloudfront.DeleteResponseHeadersPolicyInput{Id: &cfdc.rpId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete RHP %s: %v", cfdc.rpId, err)
		}
		log.Printf("deleted RHP %s\n", cfdc.rpId)
	}

	if cfdc.oacId != "" {
		x, err := cfdc.client.GetOriginAccessControl(context.TODO(), &cloudfront.GetOriginAccessControlInput{Id: &cfdc.oacId})
		if err != nil {
			log.Fatalf("could not get OAC %s: %v", cfdc.oacId, err)
		}
		_, err = cfdc.client.DeleteOriginAccessControl(context.TODO(), &cloudfront.DeleteOriginAccessControlInput{Id: &cfdc.oacId, IfMatch: x.ETag})
		if err != nil {
			log.Fatalf("could not delete OAC %s: %v", cfdc.oacId, err)
		}
		log.Printf("deleted OAC %s\n", cfdc.oacId)
	}
}

func (cfdc *distributionCreator) DisableIt() {
	distro, err := cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &cfdc.distroId})
	if err != nil {
		log.Fatalf("failed to recover distribution for %s: %v", cfdc.distroId, err)
	}

	if *distro.Distribution.Status == "Deployed" && *distro.Distribution.DistributionConfig.Enabled {
		log.Printf("Disabling %s\n", cfdc.distroId)
		isFalse := false
		distro.Distribution.DistributionConfig.Enabled = &isFalse
		_, err := cfdc.client.UpdateDistribution(context.TODO(), &cloudfront.UpdateDistributionInput{Id: &cfdc.distroId, IfMatch: distro.ETag, DistributionConfig: distro.Distribution.DistributionConfig})
		if err != nil {
			log.Fatalf("error disabling %s: %v\n", cfdc.distroId, err)
		}
	}

	for {
		distro, err := cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &cfdc.distroId})
		if err != nil {
			log.Fatalf("failed to recover distribution for %s: %v", cfdc.distroId, err)
		}
		log.Printf("disabling distro %s ... %v %s\n", cfdc.distroId, *distro.Distribution.DistributionConfig.Enabled, *distro.Distribution.Status)
		if !*distro.Distribution.DistributionConfig.Enabled && *distro.Distribution.Status != "InProgress" {
			return
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func (cfdc *distributionCreator) DeleteIt() {
	distro, err := cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &cfdc.distroId})
	if err != nil {
		log.Fatalf("failed to recover distribution for %s: %v", cfdc.distroId, err)
	}
	fmt.Printf("have a distro %s %v\n", *distro.Distribution.Status, *distro.Distribution.DistributionConfig.Enabled)

	if *distro.Distribution.DistributionConfig.Enabled {
		log.Fatalf("the distribution is still enabled")
	}
	if *distro.Distribution.Status == "Deployed" {
		log.Printf("Deleting %s\n", cfdc.distroId)
		_, err := cfdc.client.DeleteDistribution(context.TODO(), &cloudfront.DeleteDistributionInput{Id: &cfdc.distroId, IfMatch: distro.ETag})
		if err != nil {
			log.Fatalf("error deleting %s: %v\n", cfdc.distroId, err)
		}
	}

	for {
		distro, err := cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &cfdc.distroId})
		if err != nil {
			// We can't get it because it isn't there
			return
		}
		fmt.Printf("deleting distro ... %s\n", *distro.Distribution.Status)
		if *distro.Distribution.Status != "InProgress" {
			return
		}
		time.Sleep(time.Duration(1) * time.Second)
	}
}

func (cfdc *distributionCreator) ObtainMethod(name string) pluggable.Method {
	switch name {
	case "arn":
		return &arnMethod{}
	case "domainName":
		return &domainNameMethod{}
	}
	return nil
}

type arnMethod struct {
}

func (a *arnMethod) Invoke(s pluggable.RuntimeStorage, on pluggable.Expr, args []pluggable.Expr) any {
	e := on.Eval(s)
	cfdc, ok := e.(*distributionCreator)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a distribution, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if cfdc.alreadyExists {
		return cfdc.arn
	} else {
		return &DeferReadingArn{cfdc: cfdc}
	}
}

type DeferReadingArn struct {
	cfdc *distributionCreator
}

func (d *DeferReadingArn) String() string {
	if d.cfdc.arn == "" {
		panic("arn is still not set")
	}
	return d.cfdc.arn
}

type domainNameMethod struct {
}

func (a *domainNameMethod) Invoke(s pluggable.RuntimeStorage, on pluggable.Expr, args []pluggable.Expr) any {
	e := on.Eval(s)
	cfdc, ok := e.(*distributionCreator)
	if !ok {
		panic(fmt.Sprintf("domainName can only be called on a distribution, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if cfdc.alreadyExists {
		return cfdc.domainName
	} else {
		return &DeferReadingDomainName{cfdc: cfdc}
	}
}

type DeferReadingDomainName struct {
	cfdc *distributionCreator
}

func (d *DeferReadingDomainName) String() string {
	if d.cfdc.domainName == "" {
		panic("domainName is still not set")
	}
	return d.cfdc.domainName
}

var _ pluggable.HasMethods = &distributionCreator{}

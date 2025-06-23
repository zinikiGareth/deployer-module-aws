package cfront

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type distributionCreator struct {
	tools *corebottom.Tools

	loc         *errorsink.Location
	name        string
	origindns   driverbottom.Expr
	toid        driverbottom.Expr
	domains     driverbottom.List
	comment     driverbottom.Expr
	viewerCert  driverbottom.Expr
	oac         driverbottom.Expr
	cachePolicy driverbottom.Expr
	behaviors   driverbottom.List
	teardown    corebottom.TearDown

	client        *cloudfront.Client
	distroId      string
	alreadyExists bool
	arn           string
	domainName    string
}

func (cfdc *distributionCreator) Loc() *errorsink.Location {
	return cfdc.loc
}

func (cfdc *distributionCreator) ShortDescription() string {
	return "aws.CloudFront.Distribution[" + cfdc.name + "]"
}

func (cfdc *distributionCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.CloudFront.Distribution[")
	iw.AttrsWhere(cfdc)
	iw.TextAttr("named", cfdc.name)
	iw.EndAttrs()
}

func (cfdc *distributionCreator) BuildModel(pres driverbottom.ValuePresenter) {
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

	pres.Present(cfdc)
}

func (cfdc *distributionCreator) UpdateReality() {
	if cfdc.alreadyExists {
		log.Printf("distribution %s already existed for %s\n", cfdc.arn, cfdc.name)
		return
	}
	cpId, ok1 := cfdc.tools.Storage.EvalAsStringer(cfdc.cachePolicy)
	toid, ok2 := cfdc.tools.Storage.EvalAsStringer(cfdc.toid)
	if !ok1 || !ok2 {
		panic("!ok")
	}
	toidS := toid.String()
	cpIdS := cpId.String()
	dcb := types.DefaultCacheBehavior{TargetOriginId: &toidS, ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps, CachePolicyId: &cpIdS}
	origgins := cfdc.FigureOrigins(toidS)
	behaviors := cfdc.FigureCacheBehaviors()
	config := cfdc.BuildConfig(&dcb, behaviors, origgins)

	if cfdc.viewerCert != nil {
		cfdc.AttachViewerCert(config)
	}
	tagkey := "deployer-name"
	tags := types.Tags{Items: []types.Tag{{Key: &tagkey, Value: &cfdc.name}}}
	req, err := cfdc.client.CreateDistributionWithTags(context.TODO(), &cloudfront.CreateDistributionWithTagsInput{DistributionConfigWithTags: &types.DistributionConfigWithTags{DistributionConfig: config, Tags: &tags}})
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

}

func (cfdc *distributionCreator) DisableIt() {
tryAgain:
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

	utils.ExponentialBackoff(func() bool {
		distro, err := cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &cfdc.distroId})
		if err != nil {
			log.Fatalf("failed to recover distribution for %s: %v", cfdc.distroId, err)
		}
		log.Printf("disabling distro %s ... %v %s\n", cfdc.distroId, *distro.Distribution.DistributionConfig.Enabled, *distro.Distribution.Status)
		return *distro.Distribution.Status != "InProgress"
	})

	// This can fail from time to time.  If so, try all over again
	distro, err = cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &cfdc.distroId})
	if err != nil {
		log.Fatalf("failed to recover distribution for %s: %v", cfdc.distroId, err)
	}
	if *distro.Distribution.DistributionConfig.Enabled {
		goto tryAgain
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

	utils.ExponentialBackoff(func() bool {
		distro, err := cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &cfdc.distroId})
		if err != nil {
			// We can't get it because it isn't there
			return true
		}
		fmt.Printf("deleting distro ... %s\n", *distro.Distribution.Status)
		return *distro.Distribution.Status != "InProgress"
	})
}

func (cfdc *distributionCreator) FigureOrigins(targetOriginId string) *types.Origins {
	oacId, ok1 := cfdc.tools.Storage.EvalAsStringer(cfdc.oac)
	origindns, ok2 := cfdc.tools.Storage.EvalAsStringer(cfdc.origindns)
	if !ok1 || !ok2 {
		panic("!ok")
	}
	oacIdS := oacId.String()
	origindnsS := origindns.String()

	empty := ""
	s3orig := types.S3OriginConfig{OriginAccessIdentity: &empty}
	origin := types.Origin{DomainName: &origindnsS, Id: &targetOriginId, OriginAccessControlId: &oacIdS, S3OriginConfig: &s3orig}
	log.Printf("have origin %s %s\n", *origin.Id, *origin.DomainName)
	origins := []types.Origin{origin}
	nOrigins := int32(len(origins))
	return &types.Origins{Items: origins, Quantity: &nOrigins}
}

func (cfdc *distributionCreator) FigureCacheBehaviors() *types.CacheBehaviors {
	cbci := cfdc.behaviors.Eval(cfdc.tools.Storage) // TODO: expect a list
	cbcl, ok := cbci.([]any)
	if !ok {
		log.Fatalf("not a list but %T", cbci)
	}
	cbs := []types.CacheBehavior{}
	for _, m := range cbcl {
		cbc, ok := m.(cbModel)
		if !ok {
			log.Fatalf("not a cache behavior but %T", cbci)
		}
		resolved := cbc.Complete()
		log.Printf("have cb %s\n", *resolved.TargetOriginId)
		cbs = append(cbs, resolved)
	}

	cbl := int32(len(cbs))
	return &types.CacheBehaviors{Quantity: &cbl, Items: cbs}
}

func (cfdc *distributionCreator) BuildConfig(dcb *types.DefaultCacheBehavior, behaviors *types.CacheBehaviors, origins *types.Origins) *types.DistributionConfig {
	comment, ok := cfdc.tools.Storage.EvalAsStringer(cfdc.comment)
	if !ok {
		log.Fatalf("!ok, %T", comment)
	}
	commentS := comment.String()
	e := true
	dme := cfdc.tools.Storage.Eval(cfdc.domains)
	domains, ok := dme.([]any)
	if !ok {
		log.Fatalf("!ok, %T", dme)
	}
	aliases, ok := utils.AsStringList(domains)
	if !ok {
		log.Fatalf("needs to be list of strings")
	}
	nAliases := int32(len(aliases))
	rootObject := "index.html"
	return &types.DistributionConfig{CallerReference: &cfdc.name, Comment: &commentS, DefaultCacheBehavior: dcb, CacheBehaviors: behaviors, Enabled: &e, DefaultRootObject: &rootObject, Origins: origins, Aliases: &types.Aliases{Items: aliases, Quantity: &nAliases}}
}

func (cfdc *distributionCreator) AttachViewerCert(config *types.DistributionConfig) {
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

func (cfdc *distributionCreator) ObtainMethod(name string) driverbottom.Method {
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

func (a *arnMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
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
		return utils.DeferString(func() string {
			if cfdc.arn == "" {
				panic("arn is still not set")
			}
			return cfdc.arn
		})
	}
}

type domainNameMethod struct {
}

func (a *domainNameMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
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
		return utils.DeferString(func() string {
			if cfdc.domainName == "" {
				panic("domainName is still not set")
			}
			return cfdc.domainName
		})
	}
}

var _ driverbottom.HasMethods = &distributionCreator{}

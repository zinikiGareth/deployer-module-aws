package cfront

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type distributionCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *cloudfront.Client
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

func (acmc *distributionCreator) CoinId() corebottom.CoinId {
	return acmc.coin
}

func (cfdc *distributionCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
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
				model := &DistributionModel{name: cfdc.name, loc: cfdc.loc, coin: cfdc.coin /*comment: comment, origindns: src, oac: oac, behaviors: cbs, cachePolicy: cp, domains: domain, viewerCert: cert, toid: toid*/}

				model.arn = *p.ARN
				model.distroId = *p.Id
				model.domainName = *p.DomainName
				log.Printf("found distro %s: %s %s %s\n", model.name, model.arn, model.distroId, model.domainName)

				pres.Present(model)
				return
			}
		}
	}
	pres.NotFound()
}

func (cfdc *distributionCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var cert driverbottom.Expr
	var domain driverbottom.List
	var oac driverbottom.Expr
	var cbs driverbottom.List
	var cp driverbottom.Expr
	var comment driverbottom.Expr
	var src driverbottom.Expr
	var toid driverbottom.Expr
	for p, v := range cfdc.props {
		switch p.Id() {
		case "Certificate":
			cert = v
		case "OriginDNS":
			src = v
		case "Comment":
			comment = v
		case "Domain":
			le, isList := v.(driverbottom.List)
			if isList {
				domain = le
			} else {
				domain = drivertop.NewListExpr(le.Loc(), []driverbottom.Expr{v})
			}
		case "OriginAccessControl":
			oac = v
		case "CacheBehaviors":
			le, isList := v.(driverbottom.List)
			if isList {
				cbs = le
			} else {
				cbs = drivertop.NewListExpr(le.Loc(), []driverbottom.Expr{v})
			}
		case "CachePolicy":
			cp = v
		case "TargetOriginId":
			toid = v
		default:
			cfdc.tools.Reporter.ReportAtf(cfdc.loc, "invalid property for Distribution: %s", p.Id())
		}
	}
	if comment == nil {
		cfdc.tools.Reporter.ReportAtf(cfdc.loc, "Comment was not defined")
	}
	if src == nil {
		cfdc.tools.Reporter.ReportAtf(cfdc.loc, "OriginDNS was not defined")
	}
	if toid == nil {
		cfdc.tools.Reporter.ReportAtf(cfdc.loc, "TargetOriginId was not defined")
	}

	model := &DistributionModel{name: cfdc.name, loc: cfdc.loc, coin: cfdc.coin, comment: comment, origindns: src, oac: oac, behaviors: cbs, cachePolicy: cp, domains: domain, viewerCert: cert, toid: toid}
	pres.Present(model)
}

func (cfdc *distributionCreator) UpdateReality() {
	tmp := cfdc.tools.Storage.GetCoin(cfdc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*DistributionModel)
		log.Printf("distribution %s already existed for %s\n", found.arn, found.name)
		return
	}

	desired := cfdc.tools.Storage.GetCoin(cfdc.coin, corebottom.DETERMINE_DESIRED_MODE).(*DistributionModel)

	created := &DistributionModel{name: cfdc.name, loc: cfdc.loc, coin: cfdc.coin}

	cpId, ok1 := cfdc.tools.Storage.EvalAsStringer(desired.cachePolicy)
	toid, ok2 := cfdc.tools.Storage.EvalAsStringer(desired.toid)
	if !ok1 || !ok2 {
		panic("!ok")
	}
	toidS := toid.String()
	cpIdS := cpId.String()
	dcb := types.DefaultCacheBehavior{TargetOriginId: &toidS, ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps, CachePolicyId: &cpIdS}
	origins := cfdc.FigureOrigins(desired, toidS)
	behaviors := cfdc.FigureCacheBehaviors(desired)
	config := cfdc.BuildConfig(desired, &dcb, behaviors, origins)

	if desired.viewerCert != nil {
		cfdc.AttachViewerCert(desired, config)
	}
	tagkey := "deployer-name"
	tags := types.Tags{Items: []types.Tag{{Key: &tagkey, Value: &cfdc.name}}}
	req, err := cfdc.client.CreateDistributionWithTags(context.TODO(), &cloudfront.CreateDistributionWithTagsInput{DistributionConfigWithTags: &types.DistributionConfigWithTags{DistributionConfig: config, Tags: &tags}})
	if err != nil {
		log.Fatalf("failed to create distribution %s: %v\n", cfdc.name, err)
	}
	log.Printf("created distribution %s: %s %s %s\n", cfdc.name, *req.Distribution.ARN, *req.Distribution.Id, *req.Distribution.DomainName)
	created.arn = *req.Distribution.ARN
	created.distroId = *req.Distribution.Id
	created.domainName = *req.Distribution.DomainName

	cfdc.tools.Storage.Bind(cfdc.coin, created)
}

func (cfdc *distributionCreator) TearDown() {
	tmp := cfdc.tools.Storage.GetCoin(cfdc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*DistributionModel)
		log.Printf("you have asked to tear down distribution %s (id: %s, arn: %s) with mode %s\n", cfdc.name, found.distroId, found.arn, cfdc.teardown.Mode())

		cfdc.DisableIt(found)
		cfdc.DeleteIt(found)
	} else {
		log.Printf("no distribution existed for %s\n", cfdc.name)
	}

}

func (cfdc *distributionCreator) DisableIt(model *DistributionModel) {
tryAgain:
	distro, err := cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &model.distroId})
	if err != nil {
		log.Fatalf("failed to recover distribution for %s: %v", model.distroId, err)
	}

	if *distro.Distribution.Status == "Deployed" && *distro.Distribution.DistributionConfig.Enabled {
		log.Printf("Disabling %s\n", model.distroId)
		isFalse := false
		distro.Distribution.DistributionConfig.Enabled = &isFalse
		_, err := cfdc.client.UpdateDistribution(context.TODO(), &cloudfront.UpdateDistributionInput{Id: &model.distroId, IfMatch: distro.ETag, DistributionConfig: distro.Distribution.DistributionConfig})
		if err != nil {
			log.Fatalf("error disabling %s: %v\n", model.distroId, err)
		}
	}

	utils.ExponentialBackoff(func() bool {
		distro, err := cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &model.distroId})
		if err != nil {
			log.Fatalf("failed to recover distribution for %s: %v", model.distroId, err)
		}
		log.Printf("disabling distro %s ... %v %s\n", model.distroId, *distro.Distribution.DistributionConfig.Enabled, *distro.Distribution.Status)
		return *distro.Distribution.Status != "InProgress"
	})

	// This can fail from time to time.  If so, try all over again
	distro, err = cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &model.distroId})
	if err != nil {
		log.Fatalf("failed to recover distribution for %s: %v", model.distroId, err)
	}
	if *distro.Distribution.DistributionConfig.Enabled {
		goto tryAgain
	}
}

func (cfdc *distributionCreator) DeleteIt(model *DistributionModel) {
	distro, err := cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &model.distroId})
	if err != nil {
		log.Fatalf("failed to recover distribution for %s: %v", model.distroId, err)
	}
	log.Printf("have a distro %s %v\n", *distro.Distribution.Status, *distro.Distribution.DistributionConfig.Enabled)

	if *distro.Distribution.DistributionConfig.Enabled {
		log.Fatalf("the distribution is still enabled")
	}
	if *distro.Distribution.Status == "Deployed" {
		log.Printf("Deleting %s\n", model.distroId)
		_, err := cfdc.client.DeleteDistribution(context.TODO(), &cloudfront.DeleteDistributionInput{Id: &model.distroId, IfMatch: distro.ETag})
		if err != nil {
			log.Fatalf("error deleting %s: %v\n", model.distroId, err)
		}
	}

	utils.ExponentialBackoff(func() bool {
		distro, err := cfdc.client.GetDistribution(context.TODO(), &cloudfront.GetDistributionInput{Id: &model.distroId})
		if err != nil {
			// We can't get it because it isn't there
			return true
		}
		fmt.Printf("deleting distro ... %s\n", *distro.Distribution.Status)
		return *distro.Distribution.Status != "InProgress"
	})
}

func (cfdc *distributionCreator) FigureOrigins(desired *DistributionModel, targetOriginId string) *types.Origins {
	oacId, ok1 := cfdc.tools.Storage.EvalAsStringer(desired.oac)
	origindns, ok2 := cfdc.tools.Storage.EvalAsStringer(desired.origindns)
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

func (cfdc *distributionCreator) FigureCacheBehaviors(desired *DistributionModel) *types.CacheBehaviors {
	cbci := desired.behaviors.Eval(cfdc.tools.Storage)
	cbcl, ok := cbci.([]any)
	if !ok {
		log.Fatalf("not a list but %T", cbci)
	}
	cbs := []types.CacheBehavior{}
	for _, m := range cbcl {
		cbc, ok := m.(*cbModel)
		if !ok {
			log.Fatalf("not a cache behavior but %T", cbci)
		}
		resolved := cbc.Complete()
		cbs = append(cbs, resolved)
	}

	cbl := int32(len(cbs))
	return &types.CacheBehaviors{Quantity: &cbl, Items: cbs}
}

func (cfdc *distributionCreator) BuildConfig(desired *DistributionModel, dcb *types.DefaultCacheBehavior, behaviors *types.CacheBehaviors, origins *types.Origins) *types.DistributionConfig {
	comment, ok := cfdc.tools.Storage.EvalAsStringer(desired.comment)
	if !ok {
		log.Fatalf("!ok, %T", comment)
	}
	commentS := comment.String()
	e := true
	dme := cfdc.tools.Storage.Eval(desired.domains)
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

func (cfdc *distributionCreator) AttachViewerCert(desired *DistributionModel, config *types.DistributionConfig) {
	vc := cfdc.tools.Storage.Eval(desired.viewerCert)
	vcs, ok := vc.(string)
	if !ok {
		tmp, ok := vc.(fmt.Stringer)
		if !ok {
			log.Fatalf("not a string or Stringer but %T", vc)
		}
		vcs = tmp.String()
	}
	minver := types.MinimumProtocolVersionTLSv122021
	supp := types.SSLSupportMethodSniOnly
	cfdef := false
	config.ViewerCertificate = &types.ViewerCertificate{ACMCertificateArn: &vcs, MinimumProtocolVersion: minver, SSLSupportMethod: supp, CloudFrontDefaultCertificate: &cfdef}
}

var _ corebottom.Ensurable = &distributionCreator{}

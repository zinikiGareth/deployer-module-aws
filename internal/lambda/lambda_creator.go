package lambda

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
	"ziniki.org/deployer/modules/aws/internal/s3"
)

type lambdaCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *lambda.Client
}

func (lc *lambdaCreator) Loc() *errorsink.Location {
	return lc.loc
}

func (lc *lambdaCreator) ShortDescription() string {
	return "aws.Lambda.Function[" + lc.name + "]"
}

func (lc *lambdaCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.Lambda.Function[")
	iw.AttrsWhere(lc)
	iw.TextAttr("named", lc.name)
	iw.EndAttrs()
}

func (lc *lambdaCreator) CoinId() corebottom.CoinId {
	return lc.coin
}

func (lc *lambdaCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := lc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	lc.client = awsEnv.LambdaClient()

	req, err := lc.client.GetFunction(context.TODO(), &lambda.GetFunctionInput{FunctionName: &lc.name})
	if err != nil {
		if !lambdaExists(err) {
			pres.NotFound()
			return
		}
		log.Fatalf("could not recover function %s: %v\n", lc.name, err)
	}
	if req == nil {
		pres.NotFound()
		return
	}
	model := &LambdaAWSModel{name: lc.name, found: req}
	pres.Present(model)
	/*
		distros, err := lc.client.ListDistributions(context.TODO(), &cloudfront.ListDistributionsInput{})
		if err != nil {
			log.Fatalf("could not list OACs")
		}
		for _, p := range distros.DistributionList.Items {
			tags, err := lc.client.ListTagsForResource(context.TODO(), &cloudfront.ListTagsForResourceInput{Resource: p.ARN})
			if err != nil {
				log.Fatalf("error trying to obtain tags for %s\n", *p.ARN)
			}
			for _, q := range tags.Tags.Items {
				if q.Key != nil && *q.Key == "deployer-name" && q.Value != nil && *q.Value == lc.name {
					model := &DistributionModel{name: lc.name, loc: lc.loc, coin: lc.coin /*comment: comment, origindns: src, oac: oac, behaviors: cbs, cachePolicy: cp, domains: domain, viewerCert: cert, toid: toid* /}

					model.arn = *p.ARN
					model.distroId = *p.Id
					model.domainName = *p.DomainName
					for _, cb := range p.CacheBehaviors.Items {
						cpid, ok := utils.AsStringer(*cb.CachePolicyId)
						if !ok {
							panic("ugh")
						}
						cbm := &cbModel{targetOriginId: *cb.TargetOriginId, pp: *cb.PathPattern, cpId: cpid, rhp: *cb.ResponseHeadersPolicyId}
						model.foundBehaviors = append(model.foundBehaviors, cbm)
					}
					log.Printf("found distro %s: %s %s %s\n", model.name, model.arn, model.distroId, model.domainName)

					pres.Present(model)
					return
				}
			}
		}
	*/
}

func (lc *lambdaCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var runtime driverbottom.Expr
	var code *s3.S3Location
	/*
		var cert driverbottom.Expr
		var domain driverbottom.List
		var defaultRoot driverbottom.Expr
		var oac driverbottom.Expr
		var cbs driverbottom.List
		var cp driverbottom.Expr
		var comment driverbottom.Expr
		var src driverbottom.Expr
		var toid driverbottom.Expr
	*/
	for p, v := range lc.props {
		switch p.Id() {
		case "Runtime":
			runtime = v
		case "Code":
			code = v.(*s3.S3Location)
		/*
			case "Certificate":
				cert = v
			case "OriginDNS":
				src = v
			case "Comment":
				comment = v
			case "DefaultRoot":
				defaultRoot = v
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
		*/
		default:
			log.Printf("%v", v)
			lc.tools.Reporter.ReportAtf(lc.loc, "invalid property for Lambda: %s", p.Id())
		}
	}
	if code == nil {
		lc.tools.Reporter.ReportAtf(lc.loc, "Code was not defined")
	}
	if runtime == nil {
		lc.tools.Reporter.ReportAtf(lc.loc, "Runtime was not defined")
	}
	/*
		if comment == nil {
			lc.tools.Reporter.ReportAtf(lc.loc, "Comment was not defined")
		}
		if src == nil {
			lc.tools.Reporter.ReportAtf(lc.loc, "OriginDNS was not defined")
		}
		if toid == nil {
			lc.tools.Reporter.ReportAtf(lc.loc, "TargetOriginId was not defined")
		}
	*/

	model := &LambdaModel{name: lc.name, loc: lc.loc, coin: lc.coin, code: code, runtime: runtime}
	pres.Present(model)
}

func (lc *lambdaCreator) UpdateReality() {
	tmp := lc.tools.Storage.GetCoin(lc.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := lc.tools.Storage.GetCoin(lc.coin, corebottom.DETERMINE_DESIRED_MODE).(*LambdaModel)
	created := &LambdaModel{name: lc.name, loc: lc.loc, coin: lc.coin}
	/*

		var defRootObj *string = nil
		if desired.defRootExpr != nil {
			tmp, ok := lc.tools.Storage.EvalAsStringer(desired.defRootExpr)
			if !ok {
				log.Fatalf("failed to evaluate DefaultRoot")
			}
			ts := tmp.String()
			defRootObj = &ts
		}
	*/

	if tmp != nil {
		found := tmp.(*LambdaAWSModel)
		/*
			created.arn = found.arn
			created.distroId = found.distroId
			created.domainName = found.domainName
		*/
		log.Printf("lambda %s already existed for %s\n", *found.found.Configuration.FunctionArn, found.name)
		/*

			diffs := figureDiffs(lc.tools, found, desired)
			if diffs == nil {
		*/
		log.Printf("not handling diffs yet, so just adopting ...")
		lc.tools.Storage.Adopt(lc.coin, found)
		return
		/*
			} else {
				curr, err := lc.client.GetDistributionConfig(context.TODO(), &cloudfront.GetDistributionConfigInput{Id: &found.distroId})
				if err != nil {
					panic(err)
				}
				etag := curr.ETag
				curr.ETag = nil
				config := curr.DistributionConfig

				config.CacheBehaviors = diffs.apply(lc.tools, lc.client, created)
				if config.DefaultRootObject != nil && *config.DefaultRootObject != *defRootObj {
					config.DefaultRootObject = defRootObj
				}
				// TODO: should allow other things to be updated too ...

				log.Printf("updating distribution")
				_, err = lc.client.UpdateDistribution(context.TODO(), &cloudfront.UpdateDistributionInput{Id: &found.distroId, IfMatch: etag, DistributionConfig: config})
				if err != nil {
					panic(err)
				}

				lc.tools.Storage.Bind(lc.coin, created)
				return
			}
		*/
	}

	/*
		cpId, ok1 := lc.tools.Storage.EvalAsStringer(desired.cachePolicy)
		toid, ok2 := lc.tools.Storage.EvalAsStringer(desired.toid)
		if !ok1 || !ok2 {
			panic("!ok")
		}
		toidS := toid.String()
		cpIdS := cpId.String()
		dcb := types.DefaultCacheBehavior{TargetOriginId: &toidS, ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps, CachePolicyId: &cpIdS}
		origins := lc.FigureOrigins(desired, toidS)
		behaviors := lc.FigureCacheBehaviors(desired)
		config := lc.BuildConfig(desired, &dcb, behaviors, origins, defRootObj)

		if desired.viewerCert != nil {
			lc.AttachViewerCert(desired, config)
		}
		tagkey := "deployer-name"
		tags := types.Tags{Items: []types.Tag{{Key: &tagkey, Value: &lc.name}}}
	*/

	rt, ok := lc.tools.Storage.EvalAsStringer(desired.runtime)
	if !ok {
		log.Fatalf("Failed to get runtime")
	}
	var runtime types.Runtime
	switch rt.String() {
	case "go":
		runtime = types.RuntimeProvidedal2023
	}
	b1, ok := lc.tools.Storage.EvalAsStringer(desired.code.Bucket)
	if !ok {
		log.Fatalf("Failed to get bucket")
	}
	b2, ok := lc.tools.Storage.EvalAsStringer(desired.code.Key)
	if !ok {
		log.Fatalf("Failed to get key")
	}
	bucket := b1.String()
	key := b2.String()

	hackRole := "arn:aws:iam::331358773365:role/aws-ziniki-staging-Role-11P80DK9U9T9L"
	fakeHandler := "main-point"

	req, err := lc.client.CreateFunction(context.TODO(), &lambda.CreateFunctionInput{FunctionName: &lc.name, Runtime: runtime, Handler: &fakeHandler, Code: &types.FunctionCode{S3Bucket: &bucket, S3Key: &key}, Role: &hackRole})
	if err != nil {
		log.Fatalf("failed to create lambda %s: %v\n", lc.name, err)
	}
	log.Printf("created lambda %s: %s\n", lc.name, *req.FunctionArn)
	created.arn = *req.FunctionArn
	/*
		created.distroId = *req.Distribution.Id
		created.domainName = *req.Distribution.DomainName
	*/

	lc.tools.Storage.Bind(lc.coin, created)
}

func (lc *lambdaCreator) TearDown() {
	tmp := lc.tools.Storage.GetCoin(lc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*LambdaAWSModel)
		log.Printf("you have asked to tear down lambda %s with mode %s\n", found.name, lc.teardown.Mode())

		req, err := lc.client.DeleteFunction(context.TODO(), &lambda.DeleteFunctionInput{FunctionName: &found.name})
		if err != nil {
			log.Fatalf("failed to delete lambda %s: %v\n", found.name, err)
		}
		log.Printf("returned %v\n", req)
	} else {
		log.Printf("no distribution existed for %s\n", lc.name)
	}
}

func lambdaExists(err error) bool {
	if err == nil {
		return true
	}
	e1, ok := err.(*smithy.OperationError)
	if ok {
		e2, ok := e1.Err.(*http.ResponseError)
		if ok {
			if e2.ResponseError.Response.StatusCode == 404 {
				switch e4 := e2.Err.(type) {
				case *types.ResourceNotFoundException:
					return false
				default:
					log.Printf("error: %T %v", e4, e4)
					panic("what error?")
				}
			}
			log.Fatalf("error: %T %v %T %v", e2.Response.Status, e2.Response.Status, e2.ResponseError.Response.StatusCode, e2.ResponseError.Response.StatusCode)
		}
		log.Fatalf("error: %T %v", e1.Err, e1.Err)
	}
	log.Fatalf("getting lambda failed: %T %v", err, err)
	panic("failed")
}

var _ corebottom.Ensurable = &lambdaCreator{}

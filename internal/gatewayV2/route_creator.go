package gatewayV2

import (
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type routeCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	path     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *apigatewayv2.Client
}

func (rc *routeCreator) Loc() *errorsink.Location {
	return rc.loc
}

func (rc *routeCreator) ShortDescription() string {
	return "aws.gateway.Route[" + rc.path + "]"
}

func (rc *routeCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.gateway.Route[")
	iw.AttrsWhere(rc)
	iw.TextAttr("path", rc.path)
	iw.EndAttrs()
}

func (rc *routeCreator) CoinId() corebottom.CoinId {
	return rc.coin
}

func (rc *routeCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := rc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	rc.client = awsEnv.ApiGatewayV2Client()

	/*
		req, err := rc.client.GetFunction(context.TODO(), &lambda.GetFunctionInput{FunctionName: &rc.path})
		if err != nil {
			if !lambdaExists(err) {
				pres.NotFound()
				return
			}
			log.Fatalf("could not recover function %s: %v\n", rc.path, err)
		}
		if req == nil {
			pres.NotFound()
			return
		}
		model := &LambdaAWSModel{name: rc.path, config: req.Configuration}
		pres.Present(model)
	*/
}

func (rc *routeCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	/*
		var runtime driverbottom.Expr
		var handler driverbottom.Expr
		var role driverbottom.Expr
		var code *s3.S3Location
		for p, v := range rc.props {
			switch p.Id() {
			case "Runtime":
				runtime = v
			case "Code":
				code = v.(*s3.S3Location)
			case "Handler":
				handler = v
			case "Role":
				role = v
			default:
				rc.tools.Reporter.ReportAtf(rc.loc, "invalid property for Lambda: %s", p.Id())
			}
		}
		if code == nil {
			rc.tools.Reporter.ReportAtf(rc.loc, "Code was not defined")
		}
		if runtime == nil {
			rc.tools.Reporter.ReportAtf(rc.loc, "Runtime was not defined")
		}
		if role == nil {
			rc.tools.Reporter.ReportAtf(rc.loc, "Role was not defined")
		}

		model := &LambdaModel{name: rc.path, loc: rc.loc, coin: rc.coin, code: code, handler: handler, runtime: runtime, role: role}
		pres.Present(model)
	*/
}

func (rc *routeCreator) UpdateReality() {
	/*
		tmp := rc.tools.Storage.GetCoin(rc.coin, corebottom.DETERMINE_INITIAL_MODE)
		desired := rc.tools.Storage.GetCoin(rc.coin, corebottom.DETERMINE_DESIRED_MODE).(*LambdaModel)
		created := &LambdaAWSModel{name: rc.path}

		var handler string
		if desired.handler != nil {
			h, ok := rc.tools.Storage.EvalAsStringer(desired.handler)
			if !ok {
				log.Fatalf("Failed to get handler")
			}
			handler = h.String()
		}

		rt, ok := rc.tools.Storage.EvalAsStringer(desired.runtime)
		if !ok {
			log.Fatalf("Failed to get runtime")
		}
		var runtime types.Runtime
		switch rt.String() {
		case "go":
			runtime = types.RuntimeProvidedal2023
			if handler == "" {
				handler = "main-point"
			}
		default:
			for _, r := range types.RuntimeProvided.Values() {
				if string(r) == rt.String() {
					runtime = types.Runtime(rt.String())
					break
				}
			}
			if runtime == "" {
				rc.tools.Reporter.ReportAtf(rc.loc, "invalid runtime: %s", rt.String())
				return
			}
		}
		b1, ok := rc.tools.Storage.EvalAsStringer(desired.code.Bucket)
		if !ok {
			log.Fatalf("Failed to get bucket")
		}
		b2, ok := rc.tools.Storage.EvalAsStringer(desired.code.Key)
		if !ok {
			log.Fatalf("Failed to get key")
		}
		bucket := b1.String()
		key := b2.String()

		roleArn, ok := rc.tools.Storage.EvalAsStringer(desired.role)
		if !ok {
			log.Fatalf("Failed to evaluate role")
		}
		role := roleArn.String()

		if handler == "" {
			rc.tools.Reporter.ReportAtf(rc.loc, "must specify Handler for Runtime %s", rt.String())
			return
		}

		if tmp != nil {
			found := tmp.(*LambdaAWSModel)
			created.config = found.config
			log.Printf("lambda %s already existed for %s\n", *found.config.FunctionArn, found.name)
			log.Printf("not handling diffs yet; just copying ...")
			rc.tools.Storage.Bind(rc.coin, created)
			return
		}

		req, err := rc.client.CreateFunction(context.TODO(), &lambda.CreateFunctionInput{FunctionName: &rc.path, Runtime: runtime, Handler: &handler, Code: &types.FunctionCode{S3Bucket: &bucket, S3Key: &key}, Role: &role})
		if err != nil {
			log.Fatalf("failed to create lambda %s: %v\n", rc.path, err)
		}
		utils.ExponentialBackoff(func() bool {
			stat, err := rc.client.GetFunction(context.TODO(), &lambda.GetFunctionInput{FunctionName: &rc.path})
			if err != nil {
				panic(err)
			}
			if stat.Configuration.State == "Active" {
				return true
			}
			log.Printf("waiting for lambda to be active, stat = %v\n", stat.Configuration.State)
			return false
		})
		log.Printf("created lambda %s: %s\n", rc.path, *req.FunctionArn)
		created.config = &types.FunctionConfiguration{FunctionArn: req.FunctionArn}

		rc.tools.Storage.Bind(rc.coin, created)
	*/
}

func (rc *routeCreator) TearDown() {
	/*
		tmp := rc.tools.Storage.GetCoin(rc.coin, corebottom.DETERMINE_INITIAL_MODE)

		if tmp != nil {
			found := tmp.(*LambdaAWSModel)
			log.Printf("you have asked to tear down lambda %s with mode %s\n", found.name, rc.teardown.Mode())

			_, err := rc.client.DeleteFunction(context.TODO(), &lambda.DeleteFunctionInput{FunctionName: &found.name})
			if err != nil {
				log.Fatalf("failed to delete lambda %s: %v\n", found.name, err)
			}
		} else {
			log.Printf("no lambda existed for %s\n", rc.path)
		}
	*/
}

var _ corebottom.Ensurable = &routeCreator{}

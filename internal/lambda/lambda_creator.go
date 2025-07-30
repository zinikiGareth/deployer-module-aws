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
	model := &LambdaAWSModel{name: lc.name, config: req.Configuration}
	pres.Present(model)
}

func (lc *lambdaCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var runtime driverbottom.Expr
	var handler driverbottom.Expr
	var role driverbottom.Expr
	var code *s3.S3Location
	for p, v := range lc.props {
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
	if role == nil {
		lc.tools.Reporter.ReportAtf(lc.loc, "Role was not defined")
	}

	model := &LambdaModel{name: lc.name, loc: lc.loc, coin: lc.coin, code: code, handler: handler, runtime: runtime, role: role}
	pres.Present(model)
}

func (lc *lambdaCreator) UpdateReality() {
	tmp := lc.tools.Storage.GetCoin(lc.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := lc.tools.Storage.GetCoin(lc.coin, corebottom.DETERMINE_DESIRED_MODE).(*LambdaModel)
	created := &LambdaAWSModel{name: lc.name}

	var handler string
	if desired.handler != nil {
		h, ok := lc.tools.Storage.EvalAsStringer(desired.handler)
		if !ok {
			log.Fatalf("Failed to get handler")
		}
		handler = h.String()
	}

	rt, ok := lc.tools.Storage.EvalAsStringer(desired.runtime)
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
			lc.tools.Reporter.ReportAtf(lc.loc, "invalid runtime: %s", rt.String())
			return
		}
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

	roleArn, ok := lc.tools.Storage.EvalAsStringer(desired.role)
	if !ok {
		log.Fatalf("Failed to evaluate role")
	}
	role := roleArn.String()

	if handler == "" {
		lc.tools.Reporter.ReportAtf(lc.loc, "must specify Handler for Runtime %s", rt.String())
		return
	}

	if tmp != nil {
		found := tmp.(*LambdaAWSModel)
		created.config = found.config
		log.Printf("lambda %s already existed for %s\n", *found.config.FunctionArn, found.name)
		log.Printf("not handling diffs yet; just copying ...")
		lc.tools.Storage.Bind(lc.coin, created)
		return
	}

	req, err := lc.client.CreateFunction(context.TODO(), &lambda.CreateFunctionInput{FunctionName: &lc.name, Runtime: runtime, Handler: &handler, Code: &types.FunctionCode{S3Bucket: &bucket, S3Key: &key}, Role: &role})
	if err != nil {
		log.Fatalf("failed to create lambda %s: %v\n", lc.name, err)
	}
	log.Printf("created lambda %s: %s\n", lc.name, *req.FunctionArn)
	created.config = &types.FunctionConfiguration{FunctionArn: req.FunctionArn}

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
		log.Printf("no lambda existed for %s\n", lc.name)
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

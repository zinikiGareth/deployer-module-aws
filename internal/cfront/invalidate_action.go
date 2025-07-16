package cfront

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type invalidateAction struct {
	tools *corebottom.Tools
	loc   *errorsink.Location

	distribution driverbottom.Expr
	paths        driverbottom.Expr

	client *cloudfront.Client
	model  *invalidateModel
}

func (ia *invalidateAction) Loc() *errorsink.Location {
	return ia.loc
}

func (ia *invalidateAction) ShortDescription() string {
	return "InvalidateAction[" + ia.distribution.String() + "]"
}

func (ia *invalidateAction) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("InvalidateAction")
	iw.AttrsWhere(ia)
	iw.NestedAttr("distribution", ia.distribution)
	iw.EndAttrs()
}

func (ia *invalidateAction) AddAdverb(adv driverbottom.Adverb, tokens []driverbottom.Token) driverbottom.Interpreter {
	ia.tools.CoreTools.Reporter.ReportAtf(adv.Loc(), "cloudfront.invalidate does not accept adverbs")
	return drivertop.NewIgnoreInnerScope()
}

func (ia *invalidateAction) AddProperty(name driverbottom.Identifier, value driverbottom.Expr) {
	switch name.Id() {

	case "Distribution":
		if ia.distribution != nil {
			ia.tools.Reporter.Report(name.Loc().Offset, "duplicate definition of Distribution")
		}
		ia.distribution = value
	case "Paths":
		if ia.paths != nil {
			ia.tools.Reporter.Report(name.Loc().Offset, "duplicate definition of Distribution")
		}
		ia.paths = value
	default:
		ia.tools.Reporter.ReportAtf(name.Loc(), "cloudfront.invalidate does not have a parameter %s", name)
	}
}

func (ia *invalidateAction) Completed() {
}

func (ia *invalidateAction) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	if ia.distribution == nil {
		ia.tools.Reporter.ReportAtf(ia.loc, "no distribution was specified")
	}
	ia.distribution.Resolve(r)
	if ia.paths != nil {
		ia.paths.Resolve(r)
	}
	return driverbottom.MAY_BE_BOUND
}

func (ia *invalidateAction) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := ia.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	ia.client = awsEnv.CFClient()

	distroId, ok := ia.tools.Storage.EvalAsStringer(ia.distribution)
	if !ok {
		ia.tools.Reporter.ReportAtf(ia.loc, "distribution must be a stringer")
		pres.NotFound()
		return
	}

	ia.model = &invalidateModel{loc: ia.loc, distroId: distroId.String()}
	pres.Present(ia.model)
}

func (ia *invalidateAction) DetermineDesiredState(pres corebottom.ValuePresenter) {
	pres.NotFound() // not looked for - I don't think this method should even be here
}

func (ia *invalidateAction) ShouldDestroy() bool {
	return false
}

func (ia *invalidateAction) UpdateReality() {
	invalidateAt := time.Now().Format("20060102030405")
	uniqueId := fmt.Sprintf("InvalidateAt%s", invalidateAt)
	var paths []string
	if len(ia.model.paths) == 0 {
		paths = append(paths, "/*")
	} else {
		panic("unimplemented - processing of paths")
	}
	var lp int32 = int32(len(paths))
	pathObj := types.Paths{Quantity: &lp, Items: paths}
	input := cloudfront.CreateInvalidationInput{DistributionId: &ia.model.distroId, InvalidationBatch: &types.InvalidationBatch{CallerReference: &uniqueId, Paths: &pathObj}}
	out, err := ia.client.CreateInvalidation(context.TODO(), &input)
	if err != nil {
		panic(err)
	}
	log.Printf("Created Invalidation List: %s %s\n", *out.Invalidation.Id, *out.Invalidation.Status)
}

func (ia *invalidateAction) TearDown() {
}

// TODO: I think this is actually a different thing again which just wants
// DetermineInitialState and UpdateReality
var _ corebottom.RealityShifter = &invalidateAction{}

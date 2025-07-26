package iam

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type roleCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	teardown corebottom.TearDown

	client *iam.Client
}

func (r *roleCreator) Loc() *errorsink.Location {
	return r.loc
}

func (r *roleCreator) ShortDescription() string {
	return "aws.IAM.Role[" + r.name + "]"
}

func (r *roleCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.IAM.Role[")
	iw.AttrsWhere(r)
	iw.TextAttr("named", r.name)
	iw.EndAttrs()
}

func (r *roleCreator) CoinId() corebottom.CoinId {
	return r.coin
}

func (r *roleCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := r.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	r.client = awsEnv.IAMClient()

	resp, err := r.client.GetRole(context.TODO(), &iam.GetRoleInput{RoleName: &r.name})
	if err != nil {
		if !roleExists(err) {
			pres.NotFound()
			return
		}
		log.Fatalf("failed to recover role %s", r.name)
	}
	pres.Present(&RoleAWSModel{role: resp.Role})
}

func (r *roleCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	log.Printf("Need to build a role for %s\n", r.name)
}

func (r *roleCreator) UpdateReality() {
	log.Printf("Need to actually create the role for %s on AWS\n", r.name)
}

func (r *roleCreator) TearDown() {
	log.Printf("Need to delete the role for %s on AWS\n", r.name)
}

func (r *roleCreator) String() string {
	return fmt.Sprintf("EnsureRole[%s]", r.name)
}

func roleExists(err error) bool {
	if err == nil {
		return true
	}
	e1, ok := err.(*smithy.OperationError)
	if ok {
		e2, ok := e1.Err.(*http.ResponseError)
		if ok {
			if e2.ResponseError.Response.StatusCode == 404 {
				switch e4 := e2.Err.(type) {
				case *types.NoSuchEntityException:
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

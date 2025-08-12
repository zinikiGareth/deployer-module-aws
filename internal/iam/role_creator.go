package iam

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/coremod/pkg/coretop"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
	"ziniki.org/deployer/modules/aws/internal/policyjson"
)

type AcceptPolicies interface {
	AddPolicies(managed []driverbottom.Expr, inline []corebottom.PolicyActionList)
}

type roleCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	managed []driverbottom.Expr
	inline  []corebottom.PolicyActionList

	client *iam.Client
	sts    *sts.Client
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

func (r *roleCreator) AddPolicies(managed []driverbottom.Expr, inline []corebottom.PolicyActionList) {
	r.managed = managed
	r.inline = inline
}

func (r *roleCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := r.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	r.client = awsEnv.IAMClient()
	r.sts = awsEnv.STSClient()

	resp, err := r.client.GetRole(context.TODO(), &iam.GetRoleInput{RoleName: &r.name})
	if err != nil {
		if !roleExists(err) {
			pres.NotFound()
			return
		}
		log.Fatalf("failed to recover role %s", r.name)
	}

	policies, err := r.client.ListRolePolicies(context.TODO(), &iam.ListRolePoliciesInput{RoleName: &r.name})
	if err != nil {
		log.Fatalf("failed to recover role %s", r.name)
	}

	pres.Present(&RoleAWSModel{role: resp.Role, policies: policies.PolicyNames})
}

func (r *roleCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	if !utils.HasProp(r.props, "Assume") {
		r.tools.Reporter.ReportAtf(r.loc, "cannot define role without Assume")
		return
	}
	assumption := utils.FindProp(r.props, nil, "Assume")
	assumeClause := assumption.Eval(r.tools.Storage)
	var assumeList corebottom.PolicyActionList
	switch ac := assumeClause.(type) {
	case corebottom.PolicyActionList:
		assumeList = ac
	default:
		panic("not a PAL")
	}
	if utils.HasProp(r.props, "Inline") {
		switch il := utils.FindProp(r.props, nil, "Inline").(type) {
		case corebottom.PolicyActionList:
			r.inline = append(r.inline, il)
		default:
			log.Fatalf("cannot handle Inline prop %T\n", il)
		}
	}
	pres.Present(&RoleModel{name: r.name, coin: r.coin, assumption: assumeList, inline: r.inline, managed: r.managed})
}

func (r *roleCreator) UpdateReality() {
	tmp := r.tools.Storage.GetCoin(r.coin, corebottom.DETERMINE_INITIAL_MODE)
	desired := r.tools.Storage.GetCoin(r.coin, corebottom.DETERMINE_DESIRED_MODE).(*RoleModel)

	created := &RoleAWSModel{}
	if tmp == nil {
		log.Printf("creating role %s\n", r.name)
		assume := coretop.NewPolicyDocument(desired.assumption.Loc())
		desired.assumption.ApplyTo(assume)
		assumeJson, err := policyjson.BuildFrom("", assume, policyjson.AssumeRoleRules())
		if err != nil {
			log.Fatalf("could not generate policy: %v", err)
		}
		out, err := r.client.CreateRole(context.TODO(), &iam.CreateRoleInput{RoleName: &r.name, AssumeRolePolicyDocument: &assumeJson})
		if err != nil {
			log.Fatalf("failed to create role %s: %v\n", r.name, err)
		}
		created.role = out.Role
	} else {
		log.Printf("updating role %s\n", r.name)
		found := tmp.(*RoleAWSModel)
		created.role = found.role
	}

	managed := map[string]string{}
	for _, mp := range desired.managed {
		name, ok := r.tools.Storage.EvalAsStringer(mp)
		if !ok {
			panic(ok)
		}
		managed[name.String()] = "--needed--"
	}

	var marker *string
	for {
		list, err := r.client.ListPolicies(context.TODO(), &iam.ListPoliciesInput{Marker: marker})
		if err != nil {
			panic(err)
		}
		for _, mp := range list.Policies {
			if managed[*mp.PolicyName] == "--needed--" {
				managed[*mp.PolicyName] = *mp.Arn
			}
		}
		if list.Marker == nil {
			break
		}
		marker = list.Marker
	}
	for n, a := range managed {
		if a == "--needed--" {
			r.tools.Reporter.ReportAtf(r.loc, "there is no managed policy %s", n)
		}
	}
	for _, a := range managed {
		r.client.AttachRolePolicy(context.TODO(), &iam.AttachRolePolicyInput{RoleName: &r.name, PolicyArn: &a})
	}

	for k, ip := range desired.inline {
		policy := coretop.NewPolicyDocument(ip.Loc())
		ip.ApplyTo(policy)
		pname := fmt.Sprintf("%s-%d", desired.name, k)
		ps, err := policyjson.BuildFrom(strings.ReplaceAll(pname, "-", ""), policy, policyjson.StandardRules())
		if err != nil {
			log.Fatalf("could not generate policy: %v", err)
		}
		_, err = r.client.PutRolePolicy(context.TODO(), &iam.PutRolePolicyInput{RoleName: &r.name, PolicyName: &pname, PolicyDocument: &ps})
		if err != nil {
			log.Fatalf("could not generate policy: %v", err)
		}
		log.Printf("attached policy %s", pname)
	}
	r.tools.Storage.Bind(r.coin, created)
}

func (r *roleCreator) TearDown() {
	tmp := r.tools.Storage.GetCoin(r.coin, corebottom.DETERMINE_INITIAL_MODE)
	if tmp == nil {
		log.Printf("there was no role %s\n", r.name)
		return
	}

	found := tmp.(*RoleAWSModel)
	for _, p := range found.policies {
		_, err := r.client.DeleteRolePolicy(context.TODO(), &iam.DeleteRolePolicyInput{RoleName: &r.name, PolicyName: &p})
		if err != nil {
			log.Fatalf("failed to delete role policy %s %s: %v", r.name, p, err)
		}
	}
	_, err := r.client.DeleteRole(context.TODO(), &iam.DeleteRoleInput{RoleName: &r.name})
	if err != nil {
		log.Fatalf("failed to delete role %s: %v", r.name, err)
	}
	log.Printf("deleted role %s\n", r.name)
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

var _ AcceptPolicies = &roleCreator{}

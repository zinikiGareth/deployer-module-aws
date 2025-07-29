package iam

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
	"ziniki.org/deployer/modules/aws/internal/policyjson"
)

type policyCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	teardown corebottom.TearDown
	props    map[driverbottom.Identifier]driverbottom.Expr

	policyDoc corebottom.PolicyDocument

	client        *iam.Client
	alreadyExists bool
}

func (p *policyCreator) Loc() *errorsink.Location {
	return p.loc
}

func (p *policyCreator) ShortDescription() string {
	return "aws.IAM.Policy[" + p.name + "]"
}

func (p *policyCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.IAM.Policy[")
	iw.AttrsWhere(p)
	iw.TextAttr("named", p.name)
	iw.EndAttrs()
}

func (b *policyCreator) CoinId() corebottom.CoinId {
	return b.coin
}

func (p *policyCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := p.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	p.client = awsEnv.IAMClient()
}

func (p *policyCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	log.Printf("Need to find and/or build a policy model for %s\n", p.name)

	// We need to do three things here:
	// 1. Find if the policy already exists
	// 2. Generate a JSON policy document
	// 3. Determine if the policy needs creating, updating or is fine.

	// 2. Generate the JSON Policy document

	/*
		log.Printf("v = %v\n", v)
		log.Printf("storage = %v", tools.Storage)
		tools.Storage.DumpTo(os.Stdout)
	*/

	var policy driverbottom.Expr
	seenErr := false
	for prop, v := range p.props {
		switch prop.Id() {
		case "Policy":
			policy = v
		default:
			p.tools.Reporter.ReportAtf(prop.Loc(), "invalid property for IAM policy: %s", prop.Id())
		}
	}
	if !seenErr && policy == nil {
		p.tools.Reporter.ReportAtf(p.Loc(), "no Policy property was specified for %s", p.name)
	}

	loc := policy.Loc()
	pv := policy.Eval(p.tools.Storage)
	if pv == nil {
		p.tools.Reporter.ReportAtf(loc, "policy was nil")
		return
	}
	pi, ok := pv.(corebottom.PolicyDocument)
	if !ok {
		p.tools.Reporter.ReportAtf(loc, "expression was not a policy but %T %v", pv, pv)
		return
	}

	p.policyDoc = pi
}

func (p *policyCreator) UpdateReality() {
	log.Printf("Need to actually create the policy for %s on AWS\n", p.name)

	json, err := policyjson.BuildFrom(p.name, p.policyDoc)
	if err != nil {
		p.tools.Reporter.Reportf(p.loc.Offset, "error compiling JSON policy: %v", err)
	}
	log.Printf("json policy:\n====\n%s\n====\n", json)

	if p.alreadyExists {
		log.Printf("policy %s already existed\n", p.name)
		return
	}

	policy := CreatePolicy(p.client, p.name, json)
	if policy != nil {
		log.Printf("created policy %s as %s with ARN %s\n", p.name, policy.Id, policy.ARN)
	}
}

func (p *policyCreator) TearDown() {
	log.Printf("Need to delete the policy for %s on AWS\n", p.name)
}

func (p *policyCreator) String() string {
	return fmt.Sprintf("EnsurePolicy[%s:%s]", "" /* eb.env.Region */, p.name)
}

type asAWS struct {
	Name string
	Id   string
	ARN  string
}

func CreatePolicy(client *iam.Client, name string, text string) *asAWS {
	pol, err := client.CreatePolicy(context.TODO(), &iam.CreatePolicyInput{PolicyName: &name, PolicyDocument: &text})
	if err != nil {
		log.Fatalf("failed to create policy: %v", err)
	}
	return &asAWS{Name: *pol.Policy.PolicyName, Id: *pol.Policy.PolicyId, ARN: *pol.Policy.Arn}
}

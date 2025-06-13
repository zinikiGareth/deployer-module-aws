package iam

import (
	"fmt"
	"log"

	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/policyjson"
)

type policyCreator struct {
	tools *pluggable.Tools

	loc      *errorsink.Location
	name     string
	teardown pluggable.TearDown
	policy   pluggable.Expr

	// client        *s3.Client
	// alreadyExists bool
}

func (p *policyCreator) Loc() *errorsink.Location {
	return p.loc
}

func (p *policyCreator) ShortDescription() string {
	return "aws.IAM.Policy[" + p.name + "]"
}

func (p *policyCreator) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.IAM.Policy[")
	iw.AttrsWhere(p)
	iw.TextAttr("named", p.name)
	iw.EndAttrs()
}

func (p *policyCreator) BuildModel(pres pluggable.ValuePresenter) {
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
	loc := p.policy.Loc()
	pv := p.policy.Eval(p.tools.Storage)
	if pv == nil {
		p.tools.Reporter.At(loc.Line)
		p.tools.Reporter.Reportf(loc.Offset, "policy was nil")
		return
	}
	pi, ok := pv.(external.PolicyDocument)
	if !ok {
		p.tools.Reporter.At(loc.Line)
		p.tools.Reporter.Reportf(loc.Offset, "expression was not a policy but %T %v", pv, pv)
		return
	}

	json, err := policyjson.BuildFrom(p.name, pi)
	if err != nil {
		p.tools.Reporter.Reportf(p.loc.Offset, "error compiling JSON policy: %v", err)
	}
	log.Printf("json policy:\n====\n%s\n====\n", json)

	/*
		eq := p.tools.Recall.ObtainDriver("aws.AwsEnv")
		awsEnv, ok := eq.(*env.AwsEnv)
		if !ok {
			panic("could not cast env to AwsEnv")
		}

		p.client = awsEnv.S3Client()
		_, err := p.client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
			Bucket: aws.String(p.name),
		})
		if err != nil {
			var api smithy.APIError
			if e.As(err, &api) {
				// log.Printf("code: %s", api.ErrorCode())
				if api.ErrorCode() == "NotFound" {
					log.Printf("bucket does not exist: %s", p.name)
				} else {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		} else {
			log.Printf("bucket exists: %s", p.name)
			p.alreadyExists = true
		}

		// TODO: do we need to capture something here?
		pres.Present(p)
	*/
}

func (p *policyCreator) UpdateReality() {
	log.Printf("Need to actually create the policy for %s on AWS\n", p.name)

	/*
		if p.alreadyExists {
			log.Printf("bucket %s already existed\n", p.name)
			return
		}

		bucket := CreateBucket(p.client, p.name)
		if bucket != nil {
			log.Printf("created bucket %s\n", *bucket.Location)
		}
	*/
}

func (p *policyCreator) TearDown() {
	log.Printf("Need to delete the policy for %s on AWS\n", p.name)
	/*
		if !p.alreadyExists {
			log.Printf("bucket %s does not exist\n", p.name)
			return
		}
		log.Printf("you have asked to tear down bucket %s %s\n", p.name, p.teardown.Mode())
		switch p.teardown.Mode() {
		case "preserve":
			log.Printf("not deleting bucket %s because teardown mode is 'preserve'", p.name)
			// case "empty" seems like it might be a reasonable option
		case "delete":
			log.Printf("deleting bucket %s with teardown mode 'delete'", p.name)
			EmptyBucket(p.client, p.name)
			DeleteBucket(p.client, p.name)
		default:
			log.Printf("cannot handle teardown mode '%s' for bucket %s", p.teardown.Mode(), p.name)
		}
	*/
}

func (p *policyCreator) String() string {
	return fmt.Sprintf("EnsurePolicy[%s:%s]", "" /* eb.env.Region */, p.name)
}

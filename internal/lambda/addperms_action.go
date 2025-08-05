package lambda

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/coremod/pkg/coretop"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type addPermsAction struct {
	tools *corebottom.Tools
	loc   *errorsink.Location
	named driverbottom.String

	actions []corebottom.PolicyRuleAction

	client *lambda.Client
}

func (a *addPermsAction) Loc() *errorsink.Location {
	return a.loc
}

func (a *addPermsAction) ShortDescription() string {
	panic("unimplemented")
}

func (a *addPermsAction) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

func (a *addPermsAction) Attach(item any) error {
	a.actions = append(a.actions, item.(corebottom.PolicyRuleAction))
	return nil
}

func (a *addPermsAction) MakeAssign(holder driverbottom.Holder, assignTo driverbottom.Identifier, action any) any {
	panic("assignment should not be allowed here")
}

func (a *addPermsAction) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	ret := driverbottom.MAY_BE_BOUND
	for _, pra := range a.actions {
		ret = ret.Merge(pra.Resolve(r))
	}
	return ret
}

func (a *addPermsAction) DetermineInitialState(pres corebottom.ValuePresenter) {
	eq := a.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	a.client = awsEnv.LambdaClient()

	// TODO: there is some "GetPolicy" thing we can do ...
}

func (a *addPermsAction) DetermineDesiredState(pres corebottom.ValuePresenter) {
}

func (a *addPermsAction) ShouldDestroy() bool {
	return false
}

func (a *addPermsAction) UpdateReality() {
	cnt := 0
	for _, pra := range a.actions {
		doc := coretop.NewPolicyDocument(a.loc)
		pra.ApplyTo(doc)
		for _, effect := range doc.Items() {
			if effect.Effect() != "Allow" {
				panic("should be Allow")
			}
			for _, act := range effect.Actions() {
				for _, res := range effect.Resources() {
					for _, pri := range effect.Principals() {
						stmtId := fmt.Sprintf("%sSid%d", a.named, cnt)
						priName := pri.Value()
						input := &lambda.AddPermissionInput{StatementId: &stmtId, Action: &act, FunctionName: &res, Principal: &priName}
						_, err := a.client.AddPermission(context.TODO(), input)
						if err != nil {
							if !alreadyExists(err) {
								panic(err)
							}
							// } else {
							// 	log.Printf("out = %s", *out.Statement)
						}
						cnt++
					}
				}
			}
		}
	}
}

func (a *addPermsAction) TearDown() {
}

func AddLambdaPermissionsAction(tools *corebottom.Tools, loc *errorsink.Location, name driverbottom.String, actions []corebottom.PolicyRuleAction) corebottom.RealityShifter {
	return &addPermsAction{tools: tools, loc: loc, named: name, actions: actions}
}

func alreadyExists(err error) bool {
	if err == nil {
		return true
	}
	e1, ok := err.(*smithy.OperationError)
	if ok {
		e2, ok := e1.Err.(*http.ResponseError)
		if ok {
			if e2.ResponseError.Response.StatusCode == 409 {
				switch e4 := e2.Err.(type) {
				case *types.ResourceConflictException:
					return true
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

var _ corebottom.Action = &addPermsAction{}
var _ corebottom.RealityShifter = &addPermsAction{}

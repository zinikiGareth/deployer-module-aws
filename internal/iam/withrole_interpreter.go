package iam

import (
	"fmt"
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/coremod/pkg/coretop"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

type attachToPolicy struct {
	list corebottom.PolicyActionList
}

func (a attachToPolicy) Attach(item any) error {
	pra, ok := item.(corebottom.PolicyRuleAction)
	if !ok {
		return fmt.Errorf("cannot attach %T to PolicyActionList, not a PolicyRuleAction", item)
	}

	a.list.Add(pra)
	return nil
}

func (a attachToPolicy) MakeAssign(holder driverbottom.Holder, assignTo driverbottom.Identifier, action any) any {
	panic("this should not be able to happen in this context")
}

type withRoleInterpreter struct {
	tools    *driverbottom.CoreTools
	withRole *WithRole
}

func (w *withRoleInterpreter) HaveTokens(scope driverbottom.Scope, tokens []driverbottom.Token) driverbottom.Interpreter {
	if len(tokens) < 1 {
		w.tools.Reporter.Report(0, "expected policy [name]")
		return drivertop.NewIgnoreInnerScope()
	}
	verb, ok := tokens[0].(driverbottom.Identifier)
	if !ok {
		w.tools.Reporter.Report(0, "expected policy [name]")
		return drivertop.NewIgnoreInnerScope()
	}
	switch verb.Id() {
	case "policy":
		if len(tokens) > 1 {
			expr, ok := w.tools.Parser.Parse(scope, tokens[1:])
			if !ok {
				return drivertop.NewIgnoreInnerScope()
			}
			w.withRole.Managed = append(w.withRole.Managed, expr)
			return drivertop.NewDisallowInnerScope(w.tools)
		} else {
			pd := coretop.NewPolicyActionList(verb.Loc())
			w.withRole.Attach(pd)

			return drivertop.NewVerbCommandInterpreter(w.tools, attachToPolicy{list: pd}, "policy-statements", false)
		}
	default:
		w.tools.Reporter.Report(0, "expected policy [name]")
		return drivertop.NewIgnoreInnerScope()
	}
}

func (w *withRoleInterpreter) Completed() {
}

type WithRole struct {
	driverbottom.Locatable
	tools   *driverbottom.CoreTools
	name    string
	Managed []driverbottom.Expr
	Inline  []corebottom.PolicyActionList
}

func (w *WithRole) Attach(item any) error {
	// log.Printf("attaching %v of type %T", item, item)
	pra, ok := item.(corebottom.PolicyActionList)
	if !ok {
		return fmt.Errorf("cannot attach %T to WithRole, not a PolicyDocument", item)
	}

	w.Inline = append(w.Inline, pra)
	return nil
}

func (w *WithRole) MakeAssign(holder driverbottom.Holder, assignTo driverbottom.Identifier, action any) any {
	panic("this should not be able to happen in this context")
}

func (w *WithRole) Name() string {
	return w.name
}

func (w *WithRole) AddAdverb(adverb driverbottom.Adverb, args []driverbottom.Token) driverbottom.Interpreter {
	panic("unimplemented")
}

func (w *WithRole) AddProperty(name driverbottom.Identifier, expr driverbottom.Expr) {
	panic("unimplemented")
}

func (w *WithRole) Completed() {
}

func (w *WithRole) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

// This needs to return an ARN
func (w *WithRole) Eval(s driverbottom.RuntimeStorage) any {
	if !s.IsMode(corebottom.UPDATE_REALITY_MODE) {
		log.Fatalf("cannot Eval WithRole in mode %d", s.CurrentMode())
	}
	panic("this is a hack")
	// return "arn:aws:iam::331358773365:role/aws-ziniki-staging-Role-11P80DK9U9T9L"
}

func (w *WithRole) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	ret := driverbottom.MAY_BE_BOUND
	for _, m := range w.Managed {
		m.Resolve(r)
	}
	for _, ip := range w.Inline {
		ip.Resolve(r)
	}
	return ret
}

func (w *WithRole) ShortDescription() string {
	panic("unimplemented")
}

func (w *WithRole) String() string {
	panic("unimplemented")
}

func CreateWithRoleInterpreter(tools *driverbottom.CoreTools, scope driverbottom.Scope, parent driverbottom.PropertyParent, prop driverbottom.Identifier, tokens []driverbottom.Token) driverbottom.Interpreter {
	if len(tokens) == 0 {
		tools.Reporter.ReportAtf(prop.Loc(), "must specify a name for the role")
		return drivertop.NewIgnoreInnerScope()
	}
	expr, ok := tools.Parser.Parse(scope, tokens)
	if !ok {
		return drivertop.NewIgnoreInnerScope()
	}
	s, ok := tools.Storage.EvalAsStringer(expr)
	if !ok {
		tools.Reporter.ReportAtf(prop.Loc(), "role name must be a string")
		return drivertop.NewIgnoreInnerScope()
	}
	withRole := &WithRole{Locatable: prop, tools: tools, name: s.String()}
	parent.AddProperty(prop, withRole)
	return &withRoleInterpreter{tools: tools, withRole: withRole}
}

var _ driverbottom.Expr = &WithRole{}
var _ driverbottom.Interpreter = &withRoleInterpreter{}

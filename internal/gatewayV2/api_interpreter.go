package gatewayV2

import (
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

type apiInterpreter struct {
	tools *driverbottom.CoreTools
	api   *apiAction
}

func (w *apiInterpreter) HaveTokens(scope driverbottom.Scope, tokens []driverbottom.Token) driverbottom.Interpreter {
	if len(tokens) < 1 {
		w.tools.Reporter.Report(0, "expected policy [name]")
		return drivertop.NewIgnoreInnerScope()
	}
	verb, ok := tokens[0].(driverbottom.Identifier)
	if !ok {
		log.Printf("--- need to implement adverbs")
		// w.tools.Reporter.Report(0, "expected policy [name]")
		return drivertop.NewIgnoreInnerScope()
	}
	switch verb.Id() {
	case "route":
		/* WRONG!!!
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
		*/
		log.Printf("--- need to implement route")
		return drivertop.NewIgnoreInnerScope()
	case "Protocol":
		fallthrough
	case "RouteSelectionExpression":
		log.Printf("--- need to support properties")
		return drivertop.NewIgnoreInnerScope()
	default:
		w.tools.Reporter.Report(0, "expected policy [name]")
		return drivertop.NewIgnoreInnerScope()
	}
}

func (w *apiInterpreter) Completed() {
	w.api.Completed()
}

type apiRoute struct {
	driverbottom.Locatable
	name  string
}

func (w *apiRoute) Attach(item any) error {
	return nil
}

func (w *apiRoute) MakeAssign(holder driverbottom.Holder, assignTo driverbottom.Identifier, action any) any {
	panic("this should not be able to happen in this context")
}

func (w *apiRoute) Name() string {
	return w.name
}

func (w *apiRoute) AddAdverb(adverb driverbottom.Adverb, args []driverbottom.Token) driverbottom.Interpreter {
	panic("unimplemented")
}

func (w *apiRoute) AddProperty(name driverbottom.Identifier, expr driverbottom.Expr) {
	panic("unimplemented")
}

func (w *apiRoute) Completed() {
}

func (w *apiRoute) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

// This needs to return an ARN
func (w *apiRoute) Eval(s driverbottom.RuntimeStorage) any {
	if !s.IsMode(corebottom.UPDATE_REALITY_MODE) {
		log.Fatalf("cannot Eval WithRole in mode %d", s.CurrentMode())
	}
	return "arn:aws:iam::331358773365:role/aws-ziniki-staging-Role-11P80DK9U9T9L"
}

func (w *apiRoute) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	ret := driverbottom.MAY_BE_BOUND
	return ret
}

func (w *apiRoute) ShortDescription() string {
	panic("unimplemented")
}

func (w *apiRoute) String() string {
	panic("unimplemented")
}

func NewApiInterpreter(tools *driverbottom.CoreTools, scope driverbottom.Scope, parent *apiAction) driverbottom.Interpreter {
	return &apiInterpreter{tools: tools, api: parent}
}

var _ driverbottom.Expr = &apiRoute{}
var _ driverbottom.Interpreter = &apiInterpreter{}

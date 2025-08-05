package gatewayV2

import (
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
	adverb, ok := tokens[0].(driverbottom.Adverb)
	if ok {
		return w.api.AddAdverb(adverb, tokens[1:])
	}
	verb, ok := tokens[0].(driverbottom.Identifier)
	if !ok {
		w.tools.Reporter.ReportAtf(verb.Loc(), "syntax error")
		return drivertop.NewIgnoreInnerScope()
	}
	switch verb.Id() {
	case "integration":
		// is this overly restrictive? can we not allow expressions?
		if len(tokens) != 2 {
			w.tools.Reporter.ReportAtf(verb.Loc(), "integration <name>")
			return drivertop.NewIgnoreInnerScope()
		}
		name, ok := tokens[1].(driverbottom.String)
		if !ok {
			w.tools.Reporter.ReportAtf(verb.Loc(), "integration: name must be a string")
			return drivertop.NewIgnoreInnerScope()
		}
		intg := &intgConfig{name: name, props: make(map[driverbottom.Identifier]driverbottom.Expr)}
		w.api.intgs = append(w.api.intgs, intg)
		return drivertop.NewPropertiesInnerScope(w.tools, intg)
	case "route":
		// is this overly restrictive? can we not allow expressions?
		if len(tokens) != 3 {
			w.tools.Reporter.ReportAtf(verb.Loc(), "route <route> <integration>")
			return drivertop.NewIgnoreInnerScope()
		}
		routePath, ok := tokens[1].(driverbottom.String)
		if !ok {
			w.tools.Reporter.ReportAtf(verb.Loc(), "route: route must be a string")
			return drivertop.NewIgnoreInnerScope()
		}
		routeIntg, ok := tokens[2].(driverbottom.String)
		if !ok {
			w.tools.Reporter.ReportAtf(verb.Loc(), "route: integration must be a string")
			return drivertop.NewIgnoreInnerScope()
		}
		route := &routeConfig{route: routePath, integration: routeIntg}
		w.api.routes = append(w.api.routes, route)
		return drivertop.NewDisallowInnerScope(w.tools)
	case "stage":
		// is this overly restrictive? can we not allow expressions?
		if len(tokens) != 2 {
			w.tools.Reporter.ReportAtf(verb.Loc(), "stage <name>")
			return drivertop.NewIgnoreInnerScope()
		}
		name, ok := tokens[1].(driverbottom.String)
		if !ok {
			w.tools.Reporter.ReportAtf(verb.Loc(), "stage: name must be a string")
			return drivertop.NewIgnoreInnerScope()
		}
		stage := &stageConfig{name: name}
		w.api.stages = append(w.api.stages, stage)
		return drivertop.NewDisallowInnerScope(w.tools)
	case "Protocol":
		fallthrough
	case "RouteSelectionExpression":
		pis := drivertop.NewPropertiesInnerScope(w.tools, w.api)
		return pis.HaveTokens(scope, tokens)
	default:
		w.tools.Reporter.ReportAtf(verb.Loc(), "invalid configuration action: %s", verb.Id())
		return drivertop.NewIgnoreInnerScope()
	}
}

func (w *apiInterpreter) Completed() {
	for _, r := range w.api.routes {
		found := false
		want := r.integration.(driverbottom.String).Text()
		for _, i := range w.api.intgs {
			have := i.name.Text()
			if want == have {
				found = true
				break
			}
		}
		if !found {
			w.tools.Reporter.ReportAtf(r.integration.Loc(), "there is no integration defined for %s", r.integration)
		}
	}
	w.api.Completed()
}

func NewApiInterpreter(tools *driverbottom.CoreTools, scope driverbottom.Scope, parent *apiAction) driverbottom.Interpreter {
	return &apiInterpreter{tools: tools, api: parent}
}

var _ driverbottom.Interpreter = &apiInterpreter{}

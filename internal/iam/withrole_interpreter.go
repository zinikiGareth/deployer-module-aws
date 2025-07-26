package iam

import (
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type WithRole struct {
	driverbottom.Locatable
	tools *driverbottom.CoreTools
	name  string
	props map[driverbottom.Identifier]driverbottom.Expr
}

func (w *WithRole) Name() string {
	return w.name
}

func (w *WithRole) Props() map[driverbottom.Identifier]driverbottom.Expr {
	return w.props
}

func (w *WithRole) AddAdverb(adverb driverbottom.Adverb, args []driverbottom.Token) driverbottom.Interpreter {
	panic("unimplemented")
}

func (w *WithRole) AddProperty(name driverbottom.Identifier, expr driverbottom.Expr) {
	panic("unimplemented")
}

func (w *WithRole) Completed() {
	// blank := &RoleBlank{}
	// w.ens := blank.Mint(w.tools, w.Loc(), coin, w.name, w.props, td)
}

func (w *WithRole) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

// This needs to return an ARN
func (w *WithRole) Eval(s driverbottom.RuntimeStorage) any {
	if !s.IsMode(corebottom.UPDATE_REALITY_MODE) {
		log.Fatalf("cannot Eval WithRole in mode %d", s.CurrentMode())
	}
	return "arn:aws:iam::331358773365:role/aws-ziniki-staging-Role-11P80DK9U9T9L"
}

func (w *WithRole) Loc() *errorsink.Location {
	panic("unimplemented")
}

func (w *WithRole) Resolve(r driverbottom.Resolver) {
	// TODO: will need to resolve things when we have them
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
	return drivertop.NewPropertiesInnerScope(tools, withRole)
}

// var _ driverbottom.Interpreter = &S3LocationInterpreter{}

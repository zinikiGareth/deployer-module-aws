package iam

import (
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type WithRole struct {
	driverbottom.Locatable
}

func (w *WithRole) AddAdverb(adverb driverbottom.Adverb, args []driverbottom.Token) driverbottom.Interpreter {
	panic("unimplemented")
}

func (w *WithRole) AddProperty(name driverbottom.Identifier, expr driverbottom.Expr) {
	panic("unimplemented")
}

func (w *WithRole) Completed() {
	// TODO: check everything we expect has been set
}

func (w *WithRole) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

func (w *WithRole) Eval(s driverbottom.RuntimeStorage) any {
	panic("unimplemented")
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

func CreateWithRoleInterpreter(tools *driverbottom.CoreTools, parent driverbottom.PropertyParent, prop driverbottom.Identifier) driverbottom.Interpreter {
	s3loc := &WithRole{Locatable: prop}
	parent.AddProperty(prop, s3loc)
	return drivertop.NewPropertiesInnerScope(tools, s3loc)
}

// var _ driverbottom.Interpreter = &S3LocationInterpreter{}

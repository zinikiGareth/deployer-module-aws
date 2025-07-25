package iam

import (
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

/*
type S3LocationInterpreter struct {
	tools  *driverbottom.CoreTools
	parent driverbottom.PropertyParent
	prop   driverbottom.Identifier

	model []driverbottom.Expr
}

func (f *S3LocationInterpreter) HaveTokens(scope driverbottom.Scope, tokens []driverbottom.Token) driverbottom.Interpreter {
	if len(tokens) != 0 {
		f.tools.Reporter.Report(tokens[0].Loc().Offset, "specify location in properties")
		return drivertop.NewIgnoreInnerScope()
	}
	return &S3LocationAttributesInterpreter{tools: f.tools, field: field}
}

func (f *S3LocationInterpreter) Completed() {
	expr := drivertop.NewListExpr(f.prop.Loc(), f.model)
	f.parent.AddProperty(f.prop, expr)
}
*/

func CreateWithRoleInterpreter(tools *driverbottom.CoreTools, parent driverbottom.PropertyParent, prop driverbottom.Identifier) driverbottom.Interpreter {
	return drivertop.NewPropertiesInnerScope(tools, parent)
}

// var _ driverbottom.Interpreter = &S3LocationInterpreter{}

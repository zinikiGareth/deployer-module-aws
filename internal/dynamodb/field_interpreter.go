package dynamodb

import (
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

type DynamoFieldInterpreter struct {
	tools  *driverbottom.CoreTools
	parent driverbottom.PropertyParent
	prop   driverbottom.Identifier

	model []driverbottom.Expr
}

func (f *DynamoFieldInterpreter) HaveTokens(scope driverbottom.Scope, tokens []driverbottom.Token) driverbottom.Interpreter {
	if len(tokens) != 2 {
		f.tools.Reporter.Report(tokens[0].Loc().Offset, "<name> <type>")
		return drivertop.NewIgnoreInnerScope()
	}
	name, ok := tokens[0].(driverbottom.Identifier)
	if !ok {
		f.tools.Reporter.Report(tokens[0].Loc().Offset, "field name must be an Identifier")
		return drivertop.NewIgnoreInnerScope()
	}
	fty, ok := tokens[1].(driverbottom.Identifier)
	if !ok {
		f.tools.Reporter.Report(tokens[1].Loc().Offset, "field type must be an Identifier")
		return drivertop.NewIgnoreInnerScope()
	}
	field := &DynamoFieldExpr{loc: name.Loc(), name: name.Id(), ftype: fty.Id()}
	f.model = append(f.model, field)
	return &DynamoFieldAttributesInterpreter{tools: f.tools, field: field}
}

func (f *DynamoFieldInterpreter) Completed() {
	expr := drivertop.NewListExpr(f.prop.Loc(), f.model)
	f.parent.AddProperty(f.prop, expr)
}

func CreateFieldInterpreter(tools *driverbottom.CoreTools, scope driverbottom.Scope, parent driverbottom.PropertyParent, prop driverbottom.Identifier, tokens []driverbottom.Token) driverbottom.Interpreter {
	return &DynamoFieldInterpreter{tools: tools, parent: parent, prop: prop}
}

var _ driverbottom.Interpreter = &DynamoFieldInterpreter{}

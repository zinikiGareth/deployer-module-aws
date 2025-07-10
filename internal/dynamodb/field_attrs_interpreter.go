package dynamodb

import (
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

type DynamoFieldAttributesInterpreter struct {
	tools *driverbottom.CoreTools
	field *DynamoFieldExpr
}

func (d *DynamoFieldAttributesInterpreter) HaveTokens(scope driverbottom.Scope, tokens []driverbottom.Token) driverbottom.Interpreter {
	adv, ok := tokens[0].(driverbottom.Adverb)
	if !ok {
		d.tools.Reporter.Reportf(tokens[0].Loc().Offset, "invalid dynamo field attribute")
		return drivertop.NewIgnoreInnerScope()
	}
	switch adv.Name() {
	case "Key":
		if len(tokens) != 2 {
			d.tools.Reporter.Reportf(tokens[0].Loc().Offset, "@Key <type>")
			return drivertop.NewIgnoreInnerScope()
		}
		id, ok := tokens[1].(driverbottom.Identifier)
		if !ok {
			d.tools.Reporter.Reportf(tokens[0].Loc().Offset, "@Key <type> must be an identifier")
			return drivertop.NewIgnoreInnerScope()
		}
		d.field.SetKeyType(d.tools, id)
	default:
		d.tools.Reporter.Reportf(tokens[0].Loc().Offset, "invalid dynamo field attribute")
		return drivertop.NewIgnoreInnerScope()
	}
	return drivertop.NewDisallowInnerScope(d.tools)
}

func (d *DynamoFieldAttributesInterpreter) Completed() {
	// Not sure if we have anything to do here - we have already attached key type
}

var _ driverbottom.Interpreter = &DynamoFieldAttributesInterpreter{}

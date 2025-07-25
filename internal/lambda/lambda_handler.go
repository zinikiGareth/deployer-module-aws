package lambda

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

type lambdaHandler struct {
	tools *corebottom.Tools
}

func (wh *lambdaHandler) Handle(attacher driverbottom.AttachResult, scope driverbottom.Scope, tokens []driverbottom.Token) driverbottom.Interpreter {
	if len(tokens) != 2 {
		wh.tools.Reporter.Report(tokens[0].Loc().Offset, "lambda.function: function-name")
		return drivertop.NewIgnoreInnerScope()
	}

	name, ok := tokens[1].(driverbottom.String)
	if !ok {
		wh.tools.Reporter.Reportf(tokens[1].Loc().Offset, "lambda.function: instance-name must be a string, not %T", tokens[1])
		return drivertop.NewIgnoreInnerScope()
	}

	ea := &lambdaAction{tools: wh.tools, loc: tokens[0].Loc(), named: name, props: make(map[driverbottom.Identifier]driverbottom.Expr)}
	if err := attacher.Attach(ea); err != nil {
		panic(err)
	}

	return drivertop.NewPropertiesInnerScope(wh.tools.CoreTools, ea)
}

func NewLambdaFunction(tools *corebottom.Tools) driverbottom.VerbCommand {
	return &lambdaHandler{tools: tools}
}

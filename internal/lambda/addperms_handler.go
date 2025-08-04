package lambda

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

type addPermsHandler struct {
	tools *corebottom.Tools
}

func (wh *addPermsHandler) Handle(attacher driverbottom.AttachResult, scope driverbottom.Scope, tokens []driverbottom.Token) driverbottom.Interpreter {
	if len(tokens) != 2 {
		wh.tools.Reporter.Report(tokens[0].Loc().Offset, "lambda.addPermissions: statement-name")
		return drivertop.NewIgnoreInnerScope()
	}

	name, ok := tokens[1].(driverbottom.String)
	if !ok {
		wh.tools.Reporter.Reportf(tokens[1].Loc().Offset, "lambda.addPermissions: statement-name must be a string, not %T", tokens[1])
		return drivertop.NewIgnoreInnerScope()
	}

	apa := &addPermsAction{tools: wh.tools, loc: tokens[0].Loc(), named: name}
	if err := attacher.Attach(apa); err != nil {
		panic(err)
	}

	return drivertop.NewVerbCommandInterpreter(wh.tools.CoreTools, apa, "policy-statements", false)
}

func AddLambdaPermissions(tools *corebottom.Tools) driverbottom.VerbCommand {
	return &addPermsHandler{tools: tools}
}

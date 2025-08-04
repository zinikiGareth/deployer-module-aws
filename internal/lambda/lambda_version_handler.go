package lambda

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

type lambdaVersionHandler struct {
	tools *corebottom.Tools
}

func (wh *lambdaVersionHandler) Handle(attacher driverbottom.AttachResult, scope driverbottom.Scope, tokens []driverbottom.Token) driverbottom.Interpreter {
	if len(tokens) != 1 {
		wh.tools.Reporter.Report(tokens[0].Loc().Offset, "lambda.publishVersion does not take args")
		return drivertop.NewIgnoreInnerScope()
	}

	apa := &lambdaVersioner{tools: wh.tools, loc: tokens[0].Loc(), defaultPV: 1.0, props: make(map[driverbottom.Identifier]driverbottom.Expr)}
	err := attacher.Attach(apa)
	if err != nil {
		panic(err)
	}

	return drivertop.NewPropertiesInnerScope(wh.tools.CoreTools, apa)
}

func PublishVersionHandler(tools *corebottom.Tools) driverbottom.VerbCommand {
	return &lambdaVersionHandler{tools: tools}
}

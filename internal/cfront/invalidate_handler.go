package cfront

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

type invalidateHandler struct {
	tools *corebottom.Tools
}

func (wh *invalidateHandler) Handle(attacher driverbottom.AttachResult, scope driverbottom.Scope, tokens []driverbottom.Token) driverbottom.Interpreter {
	var expr driverbottom.Expr

	if len(tokens) > 1 {
		var ok bool
		expr, ok = wh.tools.Parser.Parse(scope, tokens[1:])
		if !ok {
			return drivertop.NewIgnoreInnerScope()
		}
	}

	ea := &invalidateAction{tools: wh.tools, loc: tokens[0].Loc(), distribution: expr}
	if err := attacher.Attach(ea); err != nil {
		panic(err)
	}

	return drivertop.NewPropertiesInnerScope(wh.tools.CoreTools, ea)
}

func NewInvalidateHandler(tools *corebottom.Tools) driverbottom.VerbCommand {
	return &invalidateHandler{tools: tools}
}

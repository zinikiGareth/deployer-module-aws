package cfront

import (
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type websiteAction struct {
	tools *corebottom.Tools
	loc   *errorsink.Location

	named    driverbottom.String
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown
}

func (w *websiteAction) Loc() *errorsink.Location {
	return w.loc
}

func (w *websiteAction) ShortDescription() string {
	return "WebsiteAction[" + w.named.String() + "]"
}

func (w *websiteAction) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("WebsiteAction")
	iw.AttrsWhere(w)
	iw.EndAttrs()
}

func (w *websiteAction) AddAdverb(adv driverbottom.Adverb, tokens []driverbottom.Token) driverbottom.Interpreter {
	if adv.Name() == "teardown" {
		if w.teardown != nil {
			panic("duplicate teardown")
		}
		if len(tokens) != 1 {
			panic("invalid tokens")
		}
		w.teardown = &CFS3TearDown{mode: tokens[0].(driverbottom.Identifier).Id()}

	}
	return drivertop.NewDisallowInnerScope(w.tools.CoreTools)
}

func (w *websiteAction) AddProperty(name driverbottom.Identifier, expr driverbottom.Expr) {
	log.Printf("have property %s\n", name)
}

func (w *websiteAction) Completed() {
	log.Printf("completed cloudfront from s3\n")
	// TODO: create all the ensurables here and join them together in some fashion
}

type CFS3TearDown struct {
	mode string
}

func (m *CFS3TearDown) Mode() string {
	return m.mode
}

var _ driverbottom.Describable = &websiteAction{}

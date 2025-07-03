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

	coins *websiteCoins
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

func (w *websiteAction) AddProperty(name driverbottom.Identifier, value driverbottom.Expr) {
	if name.Id() == "name" {
		if w.named != nil {
			w.tools.Reporter.Report(name.Loc().Offset, "duplicate definition of name")
			return
		}
		str, ok := value.(driverbottom.String)
		if !ok {
			w.tools.Reporter.Report(value.Loc().Offset, "name must be a string")
			return
		}
		w.named = str
	} else {
		if w.props[name] != nil {
			w.tools.Reporter.Reportf(name.Loc().Offset, "duplicate definition of %s", name.Id())
			return
		}
		w.props[name] = value
	}
}

func (w *websiteAction) Completed() {
	log.Printf("completed cloudfront from s3\n")
}

func (w *websiteAction) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	// TODO: create all the ensurables here and join them together in some fashion

	notused := w.propsMap()

	w.coins = &websiteCoins{}
	cpcoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
	discoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
	teardown := &CFS3TearDown{mode: "delete"}
	w.coins.cachePolicy = &CachePolicyCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: cpcoin, name: w.named.Text() + "-cpc", props: w.useProps(notused, "MinTTL")}
	// TODO: we should generate some of these options ourselves
	// and do so by introducing vars with field expressions
	dprops := w.useProps(notused, "Certificate", "Comment", "Domain", "TargetOriginId")
	// "CacheBehaviors", "CachePolicy", "OriginAccessControl", "OriginDNS"
	// TODO: these should be "fromCoin" expressions - a special method
	// fromCoin()
	dprops[drivertop.NewIdentifierToken(w.named.Loc(), "OriginDNS")] = drivertop.MakeString("originDNS")
	// dprops[drivertop.NewIdentifierToken(w.named.Loc(), "TargetOriginId")] = drivertop.MakeString("targetOriginId")
	w.coins.distribution = &distributionCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: discoin, name: w.named.Text(), props: dprops}

	return driverbottom.MAY_BE_BOUND
}

func (w *websiteAction) propsMap() map[string]driverbottom.Identifier {
	ret := make(map[string]driverbottom.Identifier)
	for k := range w.props {
		ret[k.Id()] = k
	}
	return ret
}

func (w *websiteAction) useProps(notused map[string]driverbottom.Identifier, which ...string) map[driverbottom.Identifier]driverbottom.Expr {
	ret := make(map[driverbottom.Identifier]driverbottom.Expr)
	log.Printf("have %v\n", w.props)
	for _, s := range which {
		log.Printf("looking for %s", s)
		for k, v := range w.props {
			if k.Id() == s {
				ret[k] = v
				break
			}
		}
		notused[s] = nil
	}
	log.Printf("passing on props: %v\n", ret)
	return ret
}

func (w *websiteAction) DetermineInitialState(pres corebottom.ValuePresenter) {
	mypres := w.newCoinPresenter()
	w.coins.cachePolicy.DetermineInitialState(mypres)
	w.coins.distribution.DetermineInitialState(mypres)
	pres.Present(mypres.distro)
}

func (w *websiteAction) DetermineDesiredState(pres corebottom.ValuePresenter) {
	mypres := w.newCoinPresenter()
	w.coins.cachePolicy.DetermineDesiredState(mypres)
	w.coins.distribution.DetermineDesiredState(mypres)
	pres.Present(mypres.distro)
}

func (w *websiteAction) UpdateReality() {
	w.coins.cachePolicy.UpdateReality()
	w.coins.distribution.UpdateReality()
}

func (w *websiteAction) TearDown() {
	w.coins.distribution.TearDown()
	w.coins.cachePolicy.TearDown()
}

type CFS3TearDown struct {
	mode string
}

func (m *CFS3TearDown) Mode() string {
	return m.mode
}

var _ corebottom.RealityShifter = &websiteAction{}

// This needs to capture all the things as they come in
// We have one of these for discovery and one for desired
// We will probably also create one for "updateReality"
// It needs some kind of OOB info to know who is presenting what, although it's possible we
// could switch on model type - let's try that first
// We need to bind them to their coin names in Storage
type coinPresenter struct {
	main   *websiteAction
	cpm    *cachePolicyModel
	distro *DistributionModel
}

func (c *coinPresenter) NotFound() {
	log.Printf("not found\n")
}

func (c *coinPresenter) Present(value any) {
	w := c.main
	switch value := value.(type) {
	case *cachePolicyModel:
		c.cpm = value
		w.tools.Storage.Bind(w.coins.cachePolicy.coin, value)
	case *DistributionModel:
		c.distro = value
		w.tools.Storage.Bind(w.coins.distribution.coin, value)
	default:
		log.Printf("need to handle present(%T %v)\n", value, value)
	}
}

func (w *websiteAction) newCoinPresenter() *coinPresenter {
	// , model: &DistributionModel{name: w.named.Text(), loc: w.loc, coin: w.coins.distribution.coin}
	return &coinPresenter{main: w}
}

var _ corebottom.ValuePresenter = &coinPresenter{}

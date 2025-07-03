package cfront

import (
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/coremod/pkg/coretop"
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
	notused := w.propsMap()

	w.coins = &websiteCoins{}
	cpcoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
	oaccoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
	rhpcoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
	cbcoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
	discoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
	teardown := &CFS3TearDown{mode: "delete"}

	getcp := coretop.MakeGetCoinMethod(w.named.Loc(), cpcoin)
	getcb := coretop.MakeGetCoinMethod(w.named.Loc(), cbcoin)

	cpcProps := w.useProps(r, notused, "MinTTL")
	w.coins.cachePolicy = &CachePolicyCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: cpcoin, name: w.named.Text() + "-cpc", props: cpcProps}

	oacOpts := make(map[driverbottom.Identifier]driverbottom.Expr)
	oacOpts[drivertop.NewIdentifierToken(w.named.Loc(), "OriginAccessControlOriginType")] = drivertop.MakeString(w.named.Loc(), "s3")
	oacOpts[drivertop.NewIdentifierToken(w.named.Loc(), "SigningBehavior")] = drivertop.MakeString(w.named.Loc(), "always")
	oacOpts[drivertop.NewIdentifierToken(w.named.Loc(), "SigningProtocol")] = drivertop.MakeString(w.named.Loc(), "sigv4")
	w.coins.originAccessControl = &OACCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: oaccoin, name: w.named.Text() + "-oac", props: oacOpts}

	// TODO: loop on CacheBehaviors
	// TODO: pull this out of the CacheBehaviors map
	rhpOpts := make(map[driverbottom.Identifier]driverbottom.Expr)
	rhpOpts[drivertop.NewIdentifierToken(w.named.Loc(), "Header")] = drivertop.MakeString(w.named.Loc(), "Content-Type")
	rhpOpts[drivertop.NewIdentifierToken(w.named.Loc(), "Value")] = drivertop.MakeString(w.named.Loc(), "text/html")
	w.coins.rhp = &RHPCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: rhpcoin, name: w.named.Text() + "-rhp", props: rhpOpts}

	cbOpts := w.useProps(r, notused, "TargetOriginId")
	cbOpts[drivertop.NewIdentifierToken(w.named.Loc(), "CachePolicy")] = coretop.MakeInvokeExpr(getcp, drivertop.NewIdentifierToken(w.named.Loc(), "id"))
	cbOpts[drivertop.NewIdentifierToken(w.named.Loc(), "PathPattern")] = drivertop.MakeString(w.named.Loc(), "*.html")
	cbOpts[drivertop.NewIdentifierToken(w.named.Loc(), "ResponseHeadersPolicy")] = drivertop.MakeString(w.named.Loc(), "text/html")
	w.coins.cb = &CacheBehaviorCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: cbcoin, name: w.named.Text() + "-cb", props: cbOpts}

	// END LOOP
	// TODO: we should generate some of these options ourselves
	// and do so by introducing vars with field expressions
	dprops := w.useProps(r, notused, "Certificate", "Comment", "Domain", "TargetOriginId")
	// "CacheBehaviors", "CachePolicy", "OriginAccessControl", "OriginDNS"
	// TODO: these should be "fromCoin" expressions - a special method
	// fromCoin()
	dprops[drivertop.NewIdentifierToken(w.named.Loc(), "CacheBehaviors")] = drivertop.NewListExpr([]driverbottom.Expr{getcb})
	dprops[drivertop.NewIdentifierToken(w.named.Loc(), "OriginDNS")] = drivertop.MakeString(w.named.Loc(), "originDNS")
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

func (w *websiteAction) useProps(r driverbottom.Resolver, notused map[string]driverbottom.Identifier, which ...string) map[driverbottom.Identifier]driverbottom.Expr {
	ret := make(map[driverbottom.Identifier]driverbottom.Expr)
	log.Printf("have %v\n", w.props)
	for _, s := range which {
		log.Printf("looking for %s", s)
		for k, v := range w.props {
			if k.Id() == s {
				ret[k] = v
				v.Resolve(r)
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
	w.coins.originAccessControl.DetermineInitialState(mypres)
	w.coins.rhp.DetermineInitialState(mypres)
	w.coins.distribution.DetermineInitialState(mypres)
	pres.Present(mypres.distro)
}

func (w *websiteAction) DetermineDesiredState(pres corebottom.ValuePresenter) {
	mypres := w.newCoinPresenter()
	w.coins.cachePolicy.DetermineDesiredState(mypres)
	w.coins.originAccessControl.DetermineDesiredState(mypres)
	w.coins.rhp.DetermineDesiredState(mypres)
	w.coins.cb.Create(mypres)
	w.coins.distribution.DetermineDesiredState(mypres)
	pres.Present(mypres.distro)
}

func (w *websiteAction) UpdateReality() {
	w.coins.cachePolicy.UpdateReality()
	w.coins.originAccessControl.UpdateReality()
	w.coins.rhp.UpdateReality()
	w.coins.distribution.UpdateReality()
}

func (w *websiteAction) TearDown() {
	w.coins.distribution.TearDown()
	w.coins.rhp.TearDown()
	w.coins.originAccessControl.TearDown()
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
	oac    *oacModel
	rhp    *rhpModel
	cbm    *cbModel
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
	case *oacModel:
		c.oac = value
		w.tools.Storage.Bind(w.coins.originAccessControl.coin, value)
	case *rhpModel:
		c.rhp = value
		w.tools.Storage.Bind(w.coins.rhp.coin, value)
	case *cbModel:
		c.cbm = value
		w.tools.Storage.Bind(w.coins.cb.coin, value)
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

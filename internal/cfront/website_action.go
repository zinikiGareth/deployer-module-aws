package cfront

import (
	"fmt"
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

	bucket         driverbottom.Expr
	coins          *websiteCoins
	policyAttacher corebottom.PolicyAttacher
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
}

func (w *websiteAction) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	notused := w.propsMap()

	bucket := w.findProp(r, notused, "Bucket")
	w.bucket = bucket

	w.coins = &websiteCoins{}
	cpcoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
	oaccoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
	discoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
	teardown := &CFS3TearDown{mode: "delete"}

	getcp := coretop.MakeGetCoinMethod(w.named.Loc(), cpcoin)
	getoac := coretop.MakeGetCoinMethod(w.named.Loc(), oaccoin)

	cpcProps := w.useProps(r, notused, "MinTTL")
	w.coins.cachePolicy = &CachePolicyCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: cpcoin, name: w.named.Text() + "-cpc", props: cpcProps}

	oacOpts := make(map[driverbottom.Identifier]driverbottom.Expr)
	oacOpts[drivertop.NewIdentifierToken(w.named.Loc(), "OriginAccessControlOriginType")] = drivertop.MakeString(w.named.Loc(), "s3")
	oacOpts[drivertop.NewIdentifierToken(w.named.Loc(), "SigningBehavior")] = drivertop.MakeString(w.named.Loc(), "always")
	oacOpts[drivertop.NewIdentifierToken(w.named.Loc(), "SigningProtocol")] = drivertop.MakeString(w.named.Loc(), "sigv4")
	w.coins.originAccessControl = &OACCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: oaccoin, name: w.named.Text() + "-oac", props: oacOpts}

	cblist := w.findProp(r, notused, "CacheBehaviors")
	cbe := w.tools.Storage.Eval(cblist)
	cbs, ok := cbe.([]any)
	if !ok {
		log.Fatalf("was %T", cbe)
	}
	cbcoins := []driverbottom.Expr{}
	for _, cb := range cbs {
		cbi := cb.(map[string]interface{})
		pp := ""
		var rhs map[string]interface{}
		subName := ""
		for k, v := range cbi {
			switch k {
			case "SubName":
				subName = v.(string)
			case "PathPattern":
				pp = v.(string)
			case "ResponseHeaders":
				rhs = v.(map[string]interface{})
			default:
				w.tools.Reporter.ReportAtf(cblist.Loc(), "No CacheBehavior parameter %s", k)
			}
		}
		if subName == "" {
			w.tools.Reporter.ReportAtf(cblist.Loc(), "CacheBehaviors requires SubName")
			continue
		}
		if rhs == nil {
			w.tools.Reporter.ReportAtf(cblist.Loc(), "CacheBehaviors requires ResponseHeaders")
			continue
		}
		if pp == "" {
			w.tools.Reporter.ReportAtf(cblist.Loc(), "CacheBehaviors requires PathPattern")
			continue
		}
		cbName := fmt.Sprintf("%s-cb-%s", w.named.Text(), subName)
		rhpName := fmt.Sprintf("%s-cb-%s-rh", w.named.Text(), subName)
		header := ""
		value := ""
		for a, b := range rhs {
			switch a {
			case "Header":
				header = b.(string)
			case "Value":
				value = b.(string)
			default:
				log.Printf("no such RH property %s\n", a)
			}
		}
		if header == "" {
			w.tools.Reporter.ReportAtf(cblist.Loc(), "ResponseHeaders requires Header")
		}
		if value == "" {
			w.tools.Reporter.ReportAtf(cblist.Loc(), "ResponseHeaders requires Value")
		}
		rhpOpts := make(map[driverbottom.Identifier]driverbottom.Expr)
		rhpOpts[drivertop.NewIdentifierToken(w.named.Loc(), "Header")] = drivertop.MakeString(w.named.Loc(), header)
		rhpOpts[drivertop.NewIdentifierToken(w.named.Loc(), "Value")] = drivertop.MakeString(w.named.Loc(), value)
		rhpcoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
		rhp := &RHPCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: rhpcoin, name: rhpName, props: rhpOpts}

		cbcoin := corebottom.CoinId(w.tools.Storage.NewObjId(w.named.Loc()))
		cbOpts := w.useProps(r, notused, "TargetOriginId")
		cbOpts[drivertop.NewIdentifierToken(w.named.Loc(), "CachePolicy")] = drivertop.MakeInvokeExpr(getcp, drivertop.NewIdentifierToken(w.named.Loc(), "id"))
		cbOpts[drivertop.NewIdentifierToken(w.named.Loc(), "PathPattern")] = drivertop.MakeString(w.named.Loc(), pp)
		getrhp := coretop.MakeGetCoinMethod(w.named.Loc(), rhp.coin)
		cbOpts[drivertop.NewIdentifierToken(w.named.Loc(), "ResponseHeadersPolicy")] = drivertop.MakeInvokeExpr(getrhp, drivertop.NewIdentifierToken(w.named.Loc(), "id"))
		w.coins.cbs = append(w.coins.cbs, &CacheBehaviorCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: cbcoin, name: cbName, props: cbOpts, rhp: rhp})
		getcb := coretop.MakeGetCoinMethod(w.named.Loc(), cbcoin)
		cbcoins = append(cbcoins, getcb)
	}

	dprops := w.useProps(r, notused, "Certificate", "Comment", "DefaultRoot", "Domain", "TargetOriginId")
	dprops[drivertop.NewIdentifierToken(w.named.Loc(), "CacheBehaviors")] = drivertop.NewListExpr(w.named.Loc(), cbcoins)
	dprops[drivertop.NewIdentifierToken(w.named.Loc(), "CachePolicy")] = drivertop.MakeInvokeExpr(getcp, drivertop.NewIdentifierToken(w.named.Loc(), "id"))
	dprops[drivertop.NewIdentifierToken(w.named.Loc(), "OriginDNS")] = drivertop.MakeInvokeExpr(bucket, drivertop.NewIdentifierToken(w.named.Loc(), "dnsName"))
	dprops[drivertop.NewIdentifierToken(w.named.Loc(), "OriginAccessControl")] = drivertop.MakeInvokeExpr(getoac, drivertop.NewIdentifierToken(w.named.Loc(), "id"))

	w.coins.distribution = &distributionCreator{tools: w.tools, teardown: teardown, loc: w.loc, coin: discoin, name: w.named.Text(), props: dprops}

	iserr := false
	for k, id := range notused {
		if id != nil {
			w.tools.Reporter.ReportAtf(id.Loc(), "no such property %s on cloudfront.distribution", k)
			iserr = true
		}
	}
	if iserr {
		return driverbottom.ERROR_OCCURRED
	}
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
	for _, s := range which {
		for k, v := range w.props {
			if k.Id() == s {
				ret[k] = v
				v.Resolve(r)
				notused[s] = nil
				break
			}
		}
	}
	return ret
}

func (w *websiteAction) findProp(r driverbottom.Resolver, notused map[string]driverbottom.Identifier, which string) driverbottom.Expr {
	for k, v := range w.props {
		if k.Id() == which {
			v.Resolve(r)
			notused[which] = nil
			return v
		}
	}
	panic("could not find " + which)
}

func (w *websiteAction) DetermineInitialState(pres corebottom.ValuePresenter) {
	mypres := w.newCoinPresenter()
	w.coins.cachePolicy.DetermineInitialState(mypres)
	w.coins.originAccessControl.DetermineInitialState(mypres)
	for _, cb := range w.coins.cbs {
		cb.rhp.DetermineInitialState(mypres)
	}
	w.coins.distribution.DetermineInitialState(mypres)
	pres.Present(mypres.distro)
}

func (w *websiteAction) DetermineDesiredState(pres corebottom.ValuePresenter) {
	mypres := w.newCoinPresenter()
	bucket := w.tools.Storage.Eval(w.bucket)
	isBucket, ok := bucket.(corebottom.PolicyAttacher)
	if !ok {
		log.Fatalf("%v evaluated to %p, which was not a policy attacher but %T", w.bucket, bucket, bucket)
	}
	w.policyAttacher = isBucket

	w.coins.cachePolicy.DetermineDesiredState(mypres)
	w.coins.originAccessControl.DetermineDesiredState(mypres)
	for _, cb := range w.coins.cbs {
		cb.rhp.DetermineDesiredState(mypres)
		cb.Create(mypres)
	}
	w.coins.distribution.DetermineDesiredState(mypres)
	pres.Present(mypres.distro)
}

func (w *websiteAction) ShouldDestroy() bool {
	return false
}

func (w *websiteAction) UpdateReality() {
	w.coins.cachePolicy.UpdateReality()
	w.coins.originAccessControl.UpdateReality()
	for _, cb := range w.coins.cbs {
		cb.rhp.UpdateReality()
	}
	w.coins.distribution.UpdateReality()
	// I think we can wait until now to build the actual policy, since its only variable is the distribution id

	w.policyAttacher.Attach(w.makePolicy())
}

func (w *websiteAction) makePolicy() corebottom.PolicyDocument {
	allResources, ok := w.tools.Storage.EvalAsStringer(drivertop.MakeInvokeExpr(w.bucket, drivertop.NewIdentifierToken(w.loc, "allResources")))
	if !ok {
		log.Fatalf("allResources failed")
	}
	ret := coretop.NewPolicyDocument(w.loc)
	item := ret.Item("Allow")
	item.Action("s3:GetObject")
	item.Resource(allResources.String())
	item.Principal(coretop.NewPrincipal("Service", "cloudfront.amazonaws.com"))

	distribution := w.tools.Storage.GetCoin(w.coins.distribution.coin, corebottom.UPDATE_REALITY_MODE).(*DistributionModel)
	expr := map[string]any{}
	expr["aws:sourceArn"] = distribution.arn
	cond := map[string]any{}
	cond["StringEquals"] = expr
	item.AMore("Condition", cond)
	return ret
}

func (w *websiteAction) TearDown() {
	w.coins.distribution.TearDown()
	for _, cb := range w.coins.cbs {
		cb.rhp.TearDown()
	}
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
	main     *websiteAction
	cpm      *cachePolicyModel
	oac      *oacModel
	rhpCount int
	cbms     []*cbModel
	distro   *DistributionModel
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
		k := c.rhpCount
		c.rhpCount++
		w.tools.Storage.Bind(w.coins.cbs[k].rhp.coin, value)
	case *cbModel:
		k := len(c.cbms)
		c.cbms = append(c.cbms, nil)
		w.tools.Storage.Bind(w.coins.cbs[k].coin, value)
	case *DistributionModel:
		c.distro = value
		w.tools.Storage.Bind(w.coins.distribution.coin, value)
	default:
		log.Fatalf("need to handle present(%T %v)\n", value, value)
	}
}

func (c *coinPresenter) WantDestruction(loc *errorsink.Location) {
	panic("need to handle website.@destroy")
}

func (w *websiteAction) newCoinPresenter() *coinPresenter {
	// , model: &DistributionModel{name: w.named.Text(), loc: w.loc, coin: w.coins.distribution.coin}
	return &coinPresenter{main: w}
}

var _ corebottom.ValuePresenter = &coinPresenter{}

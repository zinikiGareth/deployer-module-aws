package lambda

import (
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/iam"
)

type lambdaAction struct {
	tools *corebottom.Tools
	loc   *errorsink.Location

	named    driverbottom.String
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	coins *lambdaCoins
}

func (l *lambdaAction) Loc() *errorsink.Location {
	return l.loc
}

func (l *lambdaAction) ShortDescription() string {
	return "WebsiteAction[" + l.named.String() + "]"
}

func (l *lambdaAction) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("WebsiteAction")
	iw.AttrsWhere(l)
	iw.EndAttrs()
}

func (l *lambdaAction) AddAdverb(adv driverbottom.Adverb, tokens []driverbottom.Token) driverbottom.Interpreter {
	if adv.Name() == "teardown" {
		if l.teardown != nil {
			panic("duplicate teardown")
		}
		if len(tokens) != 1 {
			panic("invalid tokens")
		}
		// l.teardown = &CFS3TearDown{mode: tokens[0].(driverbottom.Identifier).Id()}
	}
	return drivertop.NewDisallowInnerScope(l.tools.CoreTools)
}

func (l *lambdaAction) AddProperty(name driverbottom.Identifier, value driverbottom.Expr) {
	if l.coins == nil {
		l.coins = &lambdaCoins{}
	}
	if name.Id() == "name" {
		if l.named != nil {
			l.tools.Reporter.Report(name.Loc().Offset, "duplicate definition of name")
			return
		}
		str, ok := value.(driverbottom.String)
		if !ok {
			l.tools.Reporter.Report(value.Loc().Offset, "name must be a string")
			return
		}
		l.named = str
	} else {
		if l.props[name] != nil {
			l.tools.Reporter.Reportf(name.Loc().Offset, "duplicate definition of %s", name.Id())
			return
		}
		switch v := value.(type) {
		case *iam.WithRole:
			log.Printf("we have a withrole which I think we need to add a coin for, etc")
			l.coins.withRole = v
		}
		l.props[name] = value
	}
}

func (l *lambdaAction) Completed() {
}

func (l *lambdaAction) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	if l.coins == nil {
		l.coins = &lambdaCoins{}
	}
	notused := utils.PropsMap(l.props)

	// runtime := utils.FindProp(r, l.props, notused, "Runtime")
	/*
		l.bucket = bucket

		cpcoin := corebottom.CoinId(l.tools.Storage.NewObjId(l.named.Loc()))
		oaccoin := corebottom.CoinId(l.tools.Storage.NewObjId(l.named.Loc()))
	*/
	lambdaCoin := corebottom.CoinId(l.tools.Storage.NewObjId(l.named.Loc()))
	teardown := &LambdaTearDown{mode: "delete"}
	/*
		getcp := coretop.MakeGetCoinMethod(l.named.Loc(), cpcoin)
		getoac := coretop.MakeGetCoinMethod(l.named.Loc(), oaccoin)

		cpcProps := l.useProps(r, notused, "MinTTL")
		l.coins.cachePolicy = &CachePolicyCreator{tools: l.tools, teardown: teardown, loc: l.loc, coin: cpcoin, name: l.named.Text() + "-cpc", props: cpcProps}

		oacOpts := make(map[driverbottom.Identifier]driverbottom.Expr)
		oacOpts[drivertop.NewIdentifierToken(l.named.Loc(), "OriginAccessControlOriginType")] = drivertop.MakeString(l.named.Loc(), "s3")
		oacOpts[drivertop.NewIdentifierToken(l.named.Loc(), "SigningBehavior")] = drivertop.MakeString(l.named.Loc(), "always")
		oacOpts[drivertop.NewIdentifierToken(l.named.Loc(), "SigningProtocol")] = drivertop.MakeString(l.named.Loc(), "sigv4")
		l.coins.originAccessControl = &OACCreator{tools: l.tools, teardown: teardown, loc: l.loc, coin: oaccoin, name: l.named.Text() + "-oac", props: oacOpts}

		cblist := l.findProp(r, notused, "CacheBehaviors")
		cbe := l.tools.Storage.Eval(cblist)
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
					l.tools.Reporter.ReportAtf(cblist.Loc(), "No CacheBehavior parameter %s", k)
				}
			}
			if subName == "" {
				l.tools.Reporter.ReportAtf(cblist.Loc(), "CacheBehaviors requires SubName")
				continue
			}
			if rhs == nil {
				l.tools.Reporter.ReportAtf(cblist.Loc(), "CacheBehaviors requires ResponseHeaders")
				continue
			}
			if pp == "" {
				l.tools.Reporter.ReportAtf(cblist.Loc(), "CacheBehaviors requires PathPattern")
				continue
			}
			cbName := fmt.Sprintf("%s-cb-%s", l.named.Text(), subName)
			rhpName := fmt.Sprintf("%s-cb-%s-rh", l.named.Text(), subName)
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
				l.tools.Reporter.ReportAtf(cblist.Loc(), "ResponseHeaders requires Header")
			}
			if value == "" {
				l.tools.Reporter.ReportAtf(cblist.Loc(), "ResponseHeaders requires Value")
			}
			rhpOpts := make(map[driverbottom.Identifier]driverbottom.Expr)
			rhpOpts[drivertop.NewIdentifierToken(l.named.Loc(), "Header")] = drivertop.MakeString(l.named.Loc(), header)
			rhpOpts[drivertop.NewIdentifierToken(l.named.Loc(), "Value")] = drivertop.MakeString(l.named.Loc(), value)
			rhpcoin := corebottom.CoinId(l.tools.Storage.NewObjId(l.named.Loc()))
			rhp := &RHPCreator{tools: l.tools, teardown: teardown, loc: l.loc, coin: rhpcoin, name: rhpName, props: rhpOpts}

			cbcoin := corebottom.CoinId(l.tools.Storage.NewObjId(l.named.Loc()))
			cbOpts := l.useProps(r, notused, "TargetOriginId")
			cbOpts[drivertop.NewIdentifierToken(l.named.Loc(), "CachePolicy")] = drivertop.MakeInvokeExpr(getcp, drivertop.NewIdentifierToken(l.named.Loc(), "id"))
			cbOpts[drivertop.NewIdentifierToken(l.named.Loc(), "PathPattern")] = drivertop.MakeString(l.named.Loc(), pp)
			getrhp := coretop.MakeGetCoinMethod(l.named.Loc(), rhp.coin)
			cbOpts[drivertop.NewIdentifierToken(l.named.Loc(), "ResponseHeadersPolicy")] = drivertop.MakeInvokeExpr(getrhp, drivertop.NewIdentifierToken(l.named.Loc(), "id"))
			l.coins.cbs = append(l.coins.cbs, &CacheBehaviorCreator{tools: l.tools, teardown: teardown, loc: l.loc, coin: cbcoin, name: cbName, props: cbOpts, rhp: rhp})
			getcb := coretop.MakeGetCoinMethod(l.named.Loc(), cbcoin)
			cbcoins = append(cbcoins, getcb)
		}

		dprops := l.useProps(r, notused, "Certificate", "Comment", "DefaultRoot", "Domain", "TargetOriginId")
		dprops[drivertop.NewIdentifierToken(l.named.Loc(), "CacheBehaviors")] = drivertop.NewListExpr(l.named.Loc(), cbcoins)
		dprops[drivertop.NewIdentifierToken(l.named.Loc(), "CachePolicy")] = drivertop.MakeInvokeExpr(getcp, drivertop.NewIdentifierToken(l.named.Loc(), "id"))
		dprops[drivertop.NewIdentifierToken(l.named.Loc(), "OriginDNS")] = drivertop.MakeInvokeExpr(bucket, drivertop.NewIdentifierToken(l.named.Loc(), "dnsName"))
		dprops[drivertop.NewIdentifierToken(l.named.Loc(), "OriginAccessControl")] = drivertop.MakeInvokeExpr(getoac, drivertop.NewIdentifierToken(l.named.Loc(), "id"))

		l.coins.distribution = &distributionCreator{tools: l.tools, teardown: teardown, loc: l.loc, coin: discoin, name: l.named.Text(), props: dprops}
	*/

	if l.coins.withRole != nil {
		roleCoin := corebottom.CoinId(l.tools.Storage.NewObjId(l.coins.withRole.Loc()))
		l.coins.roleCreator = (&iam.RoleBlank{}).Mint(l.tools, l.coins.withRole.Loc(), roleCoin, l.coins.withRole.Name(), l.coins.withRole.Props(), teardown)
	}

	funcProps := utils.UseProps(r, l.props, notused, "Code", "Handler", "Role", "Runtime")
	l.coins.lambda = &lambdaCreator{tools: l.tools, teardown: teardown, loc: l.loc, coin: lambdaCoin, name: l.named.Text(), props: funcProps}
	iserr := false
	for k, id := range notused {
		if id != nil {
			l.tools.Reporter.ReportAtf(id.Loc(), "no such property %s on lambda.function", k)
			iserr = true
		}
	}
	if iserr {
		return driverbottom.ERROR_OCCURRED
	}
	return driverbottom.MAY_BE_BOUND
}

func (l *lambdaAction) DetermineInitialState(pres corebottom.ValuePresenter) {
	mypres := l.newCoinPresenter()
	/*
		l.coins.cachePolicy.DetermineInitialState(mypres)
		l.coins.originAccessControl.DetermineInitialState(mypres)
		for _, cb := range l.coins.cbs {
			cb.rhp.DetermineInitialState(mypres)
		}
	*/
	l.coins.lambda.DetermineInitialState(mypres)
	pres.Present(mypres.lambda)
}

func (l *lambdaAction) DetermineDesiredState(pres corebottom.ValuePresenter) {
	mypres := l.newCoinPresenter()
	/*
		bucket := l.tools.Storage.Eval(l.bucket)
		isBucket, ok := bucket.(corebottom.PolicyAttacher)
		if !ok {
			log.Fatalf("%v evaluated to %p, which was not a policy attacher but %T", l.bucket, bucket, bucket)
		}
		l.policyAttacher = isBucket

		l.coins.cachePolicy.DetermineDesiredState(mypres)
		l.coins.originAccessControl.DetermineDesiredState(mypres)
		for _, cb := range l.coins.cbs {
			cb.rhp.DetermineDesiredState(mypres)
			cb.Create(mypres)
		}
	*/
	l.coins.lambda.DetermineDesiredState(mypres)
	pres.Present(mypres.lambda)
}

func (l *lambdaAction) ShouldDestroy() bool {
	return false
}

func (l *lambdaAction) UpdateReality() {
	/*
		l.coins.cachePolicy.UpdateReality()
		l.coins.originAccessControl.UpdateReality()
		for _, cb := range l.coins.cbs {
			cb.rhp.UpdateReality()
		}
	*/
	l.coins.lambda.UpdateReality()
}

func (l *lambdaAction) TearDown() {
	l.coins.lambda.TearDown()
	/*
		for _, cb := range l.coins.cbs {
			cb.rhp.TearDown()
		}
		l.coins.originAccessControl.TearDown()
		l.coins.cachePolicy.TearDown()
	*/
}

type LambdaTearDown struct {
	mode string
}

func (m *LambdaTearDown) Mode() string {
	return m.mode
}

var _ corebottom.RealityShifter = &lambdaAction{}

// This needs to capture all the things as they come in
// We have one of these for discovery and one for desired
// We will probably also create one for "updateReality"
// It needs some kind of OOB info to know who is presenting what, although it's possible we
// could switch on model type - let's try that first
// We need to bind them to their coin names in Storage
type coinPresenter struct {
	main *lambdaAction
	// cpm      *cachePolicyModel
	// oac      *oacModel
	// rhpCount int
	// cbms     []*cbModel
	lambdaFound *LambdaAWSModel
	lambda      *LambdaModel
}

func (c *coinPresenter) NotFound() {
	log.Printf("not found\n")
}

func (c *coinPresenter) Present(value any) {
	l := c.main
	switch value := value.(type) {
	/*
		case *cachePolicyModel:
			c.cpm = value
			l.tools.Storage.Bind(l.coins.cachePolicy.coin, value)
		case *oacModel:
			c.oac = value
			l.tools.Storage.Bind(l.coins.originAccessControl.coin, value)
		case *rhpModel:
			k := c.rhpCount
			c.rhpCount++
			l.tools.Storage.Bind(l.coins.cbs[k].rhp.coin, value)
		case *cbModel:
			k := len(c.cbms)
			c.cbms = append(c.cbms, nil)
			l.tools.Storage.Bind(l.coins.cbs[k].coin, value)
	*/
	case *LambdaModel:
		c.lambda = value
		l.tools.Storage.Bind(l.coins.lambda.coin, value)
	case *LambdaAWSModel:
		c.lambdaFound = value
		l.tools.Storage.Bind(l.coins.lambda.coin, value)
	default:
		log.Fatalf("need to handle present(%T %v)\n", value, value)
	}
}

func (c *coinPresenter) WantDestruction(loc *errorsink.Location) {
	panic("need to handle lambda.@destroy")
}

func (l *lambdaAction) newCoinPresenter() *coinPresenter {
	return &coinPresenter{main: l}
}

var _ corebottom.ValuePresenter = &coinPresenter{}

package gatewayV2

import (
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/coremod/pkg/coretop"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/lambda"
)

type apiAction struct {
	tools *corebottom.Tools
	loc   *errorsink.Location

	named    driverbottom.String
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	intgs  []*intgConfig
	routes []*routeConfig
	stages []*stageConfig

	creators []corebottom.BasicShifter
	coins    *apiCoins
}

func (a *apiAction) Loc() *errorsink.Location {
	return a.loc
}

func (a *apiAction) ShortDescription() string {
	return "ApiAction[" + a.named.String() + "]"
}

func (a *apiAction) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("ApiAction %s", a.named)
	iw.AttrsWhere(a)
	iw.EndAttrs()
}

func (a *apiAction) AddAdverb(adv driverbottom.Adverb, tokens []driverbottom.Token) driverbottom.Interpreter {
	if adv.Name() == "teardown" {
		if a.teardown != nil {
			panic("duplicate teardown")
		}
		if len(tokens) != 1 {
			panic("invalid tokens")
		}
		a.teardown = &ApiTearDown{mode: tokens[0].(driverbottom.Identifier).Id()}
	} else {
		a.tools.Reporter.ReportAtf(adv.Loc(), "there is no adverb %s", adv.Name())
	}
	return drivertop.NewDisallowInnerScope(a.tools.CoreTools)
}

func (a *apiAction) AddProperty(name driverbottom.Identifier, value driverbottom.Expr) {
	if name.Id() == "name" {
		if a.named != nil {
			a.tools.Reporter.Report(name.Loc().Offset, "duplicate definition of name")
			return
		}
		str, ok := value.(driverbottom.String)
		if !ok {
			a.tools.Reporter.Report(value.Loc().Offset, "name must be a string")
			return
		}
		a.named = str
	} else {
		if a.props[name] != nil {
			a.tools.Reporter.Reportf(name.Loc().Offset, "duplicate definition of %s", name.Id())
			return
		}
		a.props[name] = value
	}
}

func (a *apiAction) Completed() {
	a.coins = &apiCoins{intgs: make(map[string]*integrationCreator), routes: make(map[string]*routeCreator), stages: make(map[string]*stageCreator)}
	if a.teardown == nil {
		// a.tools.Reporter.ReportAtf(a.loc, "no teardown specified")
		log.Printf("... no teardown specified")
		a.teardown = &ApiTearDown{mode: "delete"}
		// return
	}
	notused := utils.PropsMap(a.props)
	apiCoin := corebottom.CoinId(a.tools.Storage.PendingObjId(a.named.Loc()))

	// First create the Api itself
	funcProps := utils.UseProps(a.props, notused, "Protocol", "RouteSelectionExpression")
	a.coins.api = &apiCreator{tools: a.tools, teardown: a.teardown, loc: a.loc, coin: apiCoin, name: a.named.Text(), props: funcProps}
	a.creators = append(a.creators, a.coins.api)

	// TODO: this is copied here but has the wrong things ...
	apiId := drivertop.NewIdentifierToken(a.named.Loc(), "Api")
	regionId := drivertop.NewIdentifierToken(a.named.Loc(), "Region")
	targetId := drivertop.NewIdentifierToken(a.named.Loc(), "Target")
	getApi := coretop.MakeGetCoinMethod(a.named.Loc(), a.coins.api.coin)
	arnId := drivertop.NewIdentifierToken(a.named.Loc(), "id")
	integrationId := drivertop.NewIdentifierToken(a.named.Loc(), "integrationId")
	region := drivertop.MakeString(a.named.Loc(), "us-east-1")
	invokePerm := drivertop.MakeString(a.named.Loc(), "lambda:InvokeFunction")

	for _, i := range a.intgs {
		name, ok := a.tools.Storage.EvalAsStringer(i.name)
		if !ok {
			panic("not ok")
		}

		// create the integration itself
		i.coin = corebottom.CoinId(a.tools.Storage.PendingObjId(i.name.Loc()))
		i.props[apiId] = drivertop.MakeInvokeExpr(getApi, arnId)
		i.props[regionId] = region
		ic := &integrationCreator{tools: a.tools, loc: i.name.Loc(), coin: i.coin, props: i.props}
		a.coins.intgs[name.String()] = ic
		a.creators = append(a.creators, ic)

		// add the invocation permission for the lambda
		principal := coretop.NewPolicyPrincipalAction(a.tools, i.name.Loc(), drivertop.MakeString(i.name.Loc(), "Service"), drivertop.MakeString(i.name.Loc(), "apigateway.amazonaws.com"))
		allowExection := coretop.NewPolicyAllowAction(a.tools, i.name.Loc(), []driverbottom.Expr{invokePerm}, []driverbottom.Expr{utils.FindProp(i.props, nil, "Uri")}, []corebottom.UpdatePolicyAllowAction{principal})
		alp := lambda.AddLambdaPermissionsAction(a.tools, i.name.Loc(), i.name, []corebottom.PolicyRuleAction{allowExection})
		a.creators = append(a.creators, alp)
	}

	for _, r := range a.routes {
		path, ok := a.tools.Storage.EvalAsStringer(r.route)
		if !ok {
			panic("not ok")
		}
		rcoin := corebottom.CoinId(a.tools.Storage.PendingObjId(r.route.Loc()))

		rc := &routeCreator{tools: a.tools, loc: r.route.Loc(), path: r.route.Text(), coin: rcoin, props: make(map[driverbottom.Identifier]driverbottom.Expr)}
		rc.props[apiId] = drivertop.MakeInvokeExpr(getApi, arnId)
		rc.props[targetId] = drivertop.MakeInvokeExpr(a.getIntegrationCoin(r.integration), integrationId)
		a.coins.routes[path.String()] = rc
		a.creators = append(a.creators, rc)
	}

	for _, s := range a.stages {
		name, ok := a.tools.Storage.EvalAsStringer(s.name)
		if !ok {
			panic("not ok")
		}

		scoin := corebottom.CoinId(a.tools.Storage.PendingObjId(s.name.Loc()))
		sc := &stageCreator{tools: a.tools, loc: s.name.Loc(), name: s.name.Text(), coin: scoin, props: make(map[driverbottom.Identifier]driverbottom.Expr)}
		sc.props[apiId] = drivertop.MakeInvokeExpr(getApi, arnId)
		a.coins.stages[name.String()] = sc
		a.creators = append(a.creators, sc)

		dcoin := corebottom.CoinId(a.tools.Storage.PendingObjId(s.name.Loc()))
		dc := &deploymentCreator{tools: a.tools, loc: s.name.Loc(), name: s.name.Text(), coin: dcoin, props: make(map[driverbottom.Identifier]driverbottom.Expr)}
		dc.props[apiId] = drivertop.MakeInvokeExpr(getApi, arnId)
		// a.coins.stages[name.String()] = sc
		a.creators = append(a.creators, dc)
	}

	// check all properties specified have been used
	for k, id := range notused {
		if id != nil {
			a.tools.Reporter.ReportAtf(id.Loc(), "no such property %s on api.gatewayV2", k)
		}
	}
}

func (a *apiAction) getIntegrationCoin(integration driverbottom.String) driverbottom.Expr {
	name := integration.Text()
	for _, i := range a.intgs {
		if i.name.Text() == name {
			return coretop.MakeGetCoinMethod(i.name.Loc(), i.coin)
		}
	}
	a.tools.Reporter.ReportAtf(integration.Loc(), "no integration found for %s", name)
	return integration // this is bogus but won't be executed
}

func (a *apiAction) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	ret := driverbottom.MAY_BE_BOUND
	for _, c := range a.creators {
		if rc, ok := c.(driverbottom.Resolvable); ok {
			ret = ret.Merge(rc.Resolve(r))
		}
		if cp, ok := c.(corebottom.CoinProvider); ok {
			cp.CoinId().Resolve(a.tools.Storage)
		}
	}
	return ret
}

func (a *apiAction) DetermineInitialState(pres corebottom.ValuePresenter) {
	dummy := coretop.NewDummyPresenter()
	for _, c := range a.creators {
		var mypres corebottom.ValuePresenter
		cp, ok := c.(corebottom.CoinProvider)
		if ok {
			mypres = coretop.NewCoinPresenter(a.tools.Storage, cp.CoinId(), dummy)
		} else {
			mypres = dummy
		}
		c.DetermineInitialState(mypres)
	}
	// TODO: we probably should have some kind of composite model we present
}

func (a *apiAction) DetermineDesiredState(pres corebottom.ValuePresenter) {
	dummy := coretop.NewDummyPresenter()
	for _, c := range a.creators {
		var mypres corebottom.ValuePresenter
		cp, ok := c.(corebottom.CoinProvider)
		if ok {
			mypres = coretop.NewCoinPresenter(a.tools.Storage, cp.CoinId(), dummy)
		} else {
			mypres = dummy
		}
		c.DetermineDesiredState(mypres)
	}
	// TODO: we probably should have some kind of composite model we present
}

func (a *apiAction) ShouldDestroy() bool {
	return false
}

func (a *apiAction) UpdateReality() {
	for _, c := range a.creators {
		c.UpdateReality()
	}
}

func (a *apiAction) TearDown() {
	for _, c := range a.creators {
		c.TearDown()
	}
}

type ApiTearDown struct {
	mode string
}

func (m *ApiTearDown) Mode() string {
	return m.mode
}

var _ corebottom.RealityShifter = &apiAction{}

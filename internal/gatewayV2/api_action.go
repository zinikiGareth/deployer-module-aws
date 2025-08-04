package gatewayV2

import (
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/coremod/pkg/coretop"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
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

	coins *apiCoins
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

func (a *apiAction) AddRoute( /*??*/ ) {

}

func (a *apiAction) Completed() {
	a.coins = &apiCoins{routes: make(map[string]*routeCreator)}
	if a.teardown == nil {
		// a.tools.Reporter.ReportAtf(a.loc, "no teardown specified")
		log.Printf("... no teardown specified")
		a.teardown = &ApiTearDown{mode: "delete"}
		// return
	}
	notused := utils.PropsMap(a.props)
	apiCoin := corebottom.CoinId(a.tools.Storage.PendingObjId(a.named.Loc()))

	/*
		role := utils.FindProp(a.props, notused, "Role")
		switch v := role.(type) {
		case *iam.WithRole:
			a.coins.withRole = v
			roleCoin := corebottom.CoinId(a.tools.Storage.PendingObjId(a.coins.withRole.Loc()))
			a.coins.roleCoin = roleCoin
			a.coins.roleCreator = (&iam.RoleBlank{}).Mint(a.tools, a.coins.withRole.Loc(), roleCoin, a.coins.withRole.Name(), nil, teardown)
			a.coins.roleCreator.(iam.AcceptPolicies).AddPolicies(v.Managed, v.Inline)
		}

	*/
	funcProps := utils.UseProps(a.props, notused, "Protocol", "RouteSelectionExpression")
	a.coins.api = &apiCreator{tools: a.tools, teardown: a.teardown, loc: a.loc, coin: apiCoin, name: a.named.Text(), props: funcProps}

	// I'm not quite sure how to express this yet in the file

	// this deffo wants to be done here
	props := utils.UseProps(a.props, notused, "PublishVersion", "Alias")
	apiId := drivertop.NewIdentifierToken(a.named.Loc(), "Api")
	getApi := coretop.MakeGetCoinMethod(a.named.Loc(), a.coins.api.coin)
	arnId := drivertop.NewIdentifierToken(a.named.Loc(), "id")
	props[apiId] = drivertop.MakeInvokeExpr(getApi, arnId)

	route := drivertop.MakeString(a.named.Loc(), "/*")
	routeId := drivertop.NewIdentifierToken(a.named.Loc(), "Route")
	props[routeId] = route

	a.coins.routes["/*"] = &routeCreator{tools: a.tools, teardown: a.teardown, loc: a.loc}

	// check all properties specified have been used
	for k, id := range notused {
		if id != nil {
			a.tools.Reporter.ReportAtf(id.Loc(), "no such property %s on api.gatewayV2", k)
		}
	}
}

func (a *apiAction) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	a.coins.api.coin.Resolve(a.tools.Storage)
	/*
		a.coins.roleCoin.Resolve(a.tools.Storage)
		a.coins.withRole.Resolve(r)
		if a.coins.versioner != nil {
			a.coins.versioner.coin.Resolve(a.tools.Storage)
			a.coins.versioner.Resolve(r)
		}
	*/
	return driverbottom.MAY_BE_BOUND
}

func (a *apiAction) DetermineInitialState(pres corebottom.ValuePresenter) {
	mypres := a.newCoinPresenter()
	a.coins.api.DetermineInitialState(mypres)
	/*
		if a.coins.roleCreator != nil {
			a.coins.roleCreator.DetermineInitialState(mypres)
		}
		if a.coins.versioner != nil {
			a.coins.versioner.DetermineInitialState(mypres)
		}
	*/
	pres.Present(mypres)
}

func (a *apiAction) DetermineDesiredState(pres corebottom.ValuePresenter) {
	mypres := a.newCoinPresenter()
	a.coins.api.DetermineDesiredState(mypres)
	/*
		if a.coins.roleCreator != nil {
			a.coins.roleCreator.DetermineDesiredState(mypres)
		}
		if a.coins.versioner != nil {
			a.coins.versioner.DetermineDesiredState(mypres)
		}
	*/
	pres.Present(mypres)
}

func (a *apiAction) ShouldDestroy() bool {
	return false
}

func (a *apiAction) UpdateReality() {
	a.coins.api.UpdateReality()
	/*
		if a.coins.roleCreator != nil {
			a.coins.roleCreator.UpdateReality()
		}
		a.coins.lambda.UpdateReality()
		if a.coins.versioner != nil {
			a.coins.versioner.UpdateReality()
		}
	*/
}

func (a *apiAction) TearDown() {
	a.coins.api.TearDown()
	/*
		if a.coins.versioner != nil {
			a.coins.versioner.TearDown()
		}
		if a.coins.roleCreator != nil {
			a.coins.roleCreator.TearDown()
		}
	*/
}

type ApiTearDown struct {
	mode string
}

func (m *ApiTearDown) Mode() string {
	return m.mode
}

var _ corebottom.RealityShifter = &apiAction{}

type coinPresenter struct {
	main *apiAction
	// roleFound      *iam.RoleAWSModel
	// role           *iam.RoleModel
	apiFound *ApiAWSModel
	api      *ApiModel
	// publishAWS     *publishVersionAWS
	// publishVersion *publishVersionModel
}

func (c *coinPresenter) NotFound() {
	log.Printf("not found\n")
}

func (c *coinPresenter) Present(value any) {
	a := c.main
	switch value := value.(type) {
	// case *iam.RoleAWSModel:
	// 	c.roleFound = value
	// 	l.tools.Storage.Bind(l.coins.roleCoin, value)
	// case *iam.RoleModel:
	// 	c.role = value
	// 	l.tools.Storage.Bind(l.coins.roleCoin, value)
	case *ApiAWSModel:
		c.apiFound = value
		a.tools.Storage.Bind(a.coins.api.coin, value)
	case *ApiModel:
		c.api = value
		a.tools.Storage.Bind(a.coins.api.coin, value)
	// case *publishVersionAWS:
	// 	c.publishAWS = value
	// 	l.tools.Storage.Bind(l.coins.versioner.coin, value)
	// case *publishVersionModel:
	// 	c.publishVersion = value
	// 	l.tools.Storage.Bind(l.coins.versioner.coin, value)
	default:
		log.Fatalf("need to handle present(%T %v)\n", value, value)
	}
}

func (c *coinPresenter) WantDestruction(loc *errorsink.Location) {
	panic("need to handle lambda.@destroy")
}

func (a *apiAction) newCoinPresenter() *coinPresenter {
	return &coinPresenter{main: a}
}

var _ corebottom.ValuePresenter = &coinPresenter{}

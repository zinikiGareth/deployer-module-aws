package lambda

import (
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/coremod/pkg/coretop"
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
	return "LambdaAction[" + l.named.String() + "]"
}

func (l *lambdaAction) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("LambdaAction %s", l.named)
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
		l.teardown = &LambdaTearDown{mode: tokens[0].(driverbottom.Identifier).Id()}
	} else {
		l.tools.Reporter.ReportAtf(adv.Loc(), "there is no adverb %s", adv.Name())
	}
	return drivertop.NewDisallowInnerScope(l.tools.CoreTools)
}

func (l *lambdaAction) AddProperty(name driverbottom.Identifier, value driverbottom.Expr) {
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
		l.props[name] = value
	}
}

func (l *lambdaAction) Completed() {
	if l.teardown == nil {
		l.tools.Reporter.ReportAtf(l.loc, "no teardown specified")
		return
	}

	l.coins = &lambdaCoins{}
	notused := utils.PropsMap(l.props)

	lambdaCoin := corebottom.CoinId(l.tools.Storage.PendingObjId(l.named.Loc()))

	role := utils.FindProp(l.props, notused, "Role")
	funcProps := utils.UseProps(l.props, notused, "Code", "Handler", "Runtime", "VpcConfig")
	switch v := role.(type) {
	case *iam.WithRole:
		l.coins.withRole = v
		roleCoin := corebottom.CoinId(l.tools.Storage.PendingObjId(l.coins.withRole.Loc()))
		l.coins.roleCoin = roleCoin
		l.coins.roleCreator = (&iam.RoleBlank{}).Mint(l.tools, l.coins.withRole.Loc(), roleCoin, l.coins.withRole.Name(), nil, l.teardown)
		l.coins.roleCreator.(iam.AcceptPolicies).AddPolicies(v.Managed, v.Inline)
		roleId := utils.PropId(l.props, "Role")
		funcProps[roleId] = drivertop.MakeInvokeExpr(coretop.MakeGetCoinMethod(v.Loc(), roleCoin), drivertop.NewIdentifierToken(v.Loc(), "arn"))
	default:
		utils.CopyProps(funcProps, l.props, notused, "Role")
	}

	l.coins.lambda = &lambdaCreator{tools: l.tools, teardown: l.teardown, loc: l.loc, coin: lambdaCoin, name: l.named.Text(), props: funcProps}

	if utils.HasProp(l.props, "PublishVersion", "Alias") {
		versionerCoin := corebottom.CoinId(l.tools.Storage.PendingObjId(l.loc))
		props := utils.UseProps(l.props, notused, "PublishVersion", "Alias")
		nameId := drivertop.NewIdentifierToken(l.named.Loc(), "Name")
		getLambda := coretop.MakeGetCoinMethod(l.named.Loc(), l.coins.lambda.coin)
		arnId := drivertop.NewIdentifierToken(l.named.Loc(), "arn")
		props[nameId] = drivertop.MakeInvokeExpr(getLambda, arnId)
		l.coins.versioner = &lambdaVersioner{tools: l.tools, coin: versionerCoin, props: props}
	}

	// check all properties specified have been used
	for k, id := range notused {
		if id != nil {
			l.tools.Reporter.ReportAtf(id.Loc(), "no such property %s on lambda.function", k)
		}
	}
}

func (l *lambdaAction) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	ret := driverbottom.MAY_BE_BOUND
	l.coins.lambda.coin.Resolve(l.tools.Storage)
	l.coins.roleCoin.Resolve(l.tools.Storage)
	l.coins.withRole.Resolve(r)
	if l.coins.versioner != nil {
		l.coins.versioner.coin.Resolve(l.tools.Storage)
		l.coins.versioner.Resolve(r)
	}
	for _, p := range l.props {
		ret = ret.Merge(p.Resolve(r))
	}
	return ret
}

func (l *lambdaAction) DetermineInitialState(pres corebottom.ValuePresenter) {
	mypres := l.newCoinPresenter()
	if l.coins.roleCreator != nil {
		l.coins.roleCreator.DetermineInitialState(mypres)
	}
	l.coins.lambda.DetermineInitialState(mypres)
	if l.coins.versioner != nil {
		l.coins.versioner.DetermineInitialState(mypres)
	}
	pres.Present(mypres.lambda)
}

func (l *lambdaAction) DetermineDesiredState(pres corebottom.ValuePresenter) {
	mypres := l.newCoinPresenter()
	if l.coins.roleCreator != nil {
		l.coins.roleCreator.DetermineDesiredState(mypres)
	}
	l.coins.lambda.DetermineDesiredState(mypres)
	if l.coins.versioner != nil {
		l.coins.versioner.DetermineDesiredState(mypres)
	}
	pres.Present(mypres.lambda)
}

func (l *lambdaAction) ShouldDestroy() bool {
	return false
}

func (l *lambdaAction) UpdateReality() {
	if l.coins.roleCreator != nil {
		l.coins.roleCreator.UpdateReality()
	}
	l.coins.lambda.UpdateReality()
	if l.coins.versioner != nil {
		l.coins.versioner.UpdateReality()
	}
}

func (l *lambdaAction) TearDown() {
	if l.coins.versioner != nil {
		l.coins.versioner.TearDown()
	}
	l.coins.lambda.TearDown()
	if l.coins.roleCreator != nil {
		l.coins.roleCreator.TearDown()
	}
}

type LambdaTearDown struct {
	mode string
}

func (m *LambdaTearDown) Mode() string {
	return m.mode
}

var _ corebottom.RealityShifter = &lambdaAction{}

type coinPresenter struct {
	main           *lambdaAction
	roleFound      *iam.RoleAWSModel
	role           *iam.RoleModel
	lambdaFound    *LambdaAWSModel
	lambda         *LambdaModel
	publishAWS     *publishVersionAWS
	publishVersion *publishVersionModel
}

func (c *coinPresenter) NotFound() {
	log.Printf("not found\n")
}

func (c *coinPresenter) Present(value any) {
	l := c.main
	switch value := value.(type) {
	case *iam.RoleAWSModel:
		c.roleFound = value
		l.tools.Storage.Bind(l.coins.roleCoin, value)
	case *iam.RoleModel:
		c.role = value
		l.tools.Storage.Bind(l.coins.roleCoin, value)
	case *LambdaAWSModel:
		c.lambdaFound = value
		l.tools.Storage.Bind(l.coins.lambda.coin, value)
	case *LambdaModel:
		c.lambda = value
		l.tools.Storage.Bind(l.coins.lambda.coin, value)
	case *publishVersionAWS:
		c.publishAWS = value
		l.tools.Storage.Bind(l.coins.versioner.coin, value)
	case *publishVersionModel:
		c.publishVersion = value
		l.tools.Storage.Bind(l.coins.versioner.coin, value)
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

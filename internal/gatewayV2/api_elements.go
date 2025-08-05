package gatewayV2

import (
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
)

type intgConfig struct {
	name  driverbottom.String
	coin  corebottom.CoinId
	props map[driverbottom.Identifier]driverbottom.Expr
}

func (i *intgConfig) AddAdverb(adverb driverbottom.Adverb, args []driverbottom.Token) driverbottom.Interpreter {
	panic("unimplemented")
}

func (i *intgConfig) AddProperty(name driverbottom.Identifier, expr driverbottom.Expr) {
	i.props[name] = expr
}

func (i *intgConfig) Completed() {
}

type routeConfig struct {
	route       driverbottom.Expr
	integration driverbottom.String
}

type stageConfig struct {
	name driverbottom.String
}

var _ driverbottom.PropertyParent = &intgConfig{}

package cfront

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type cachePolicyModel struct {
	loc           *errorsink.Location
	name          string
	coin          corebottom.CoinId
	minttl        driverbottom.Expr
	which         string
	CachePolicyId string
}

func NewCachePolicyModel(coin corebottom.CoinId, loc *errorsink.Location, name string, which string) *cachePolicyModel {
	return &cachePolicyModel{coin: coin, loc: loc, name: name, which: which}
}

func (model *cachePolicyModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "id":
		return &CachePolicyIdMethod{}
	}
	return nil
}

type CachePolicyIdMethod struct {
}

func (a *CachePolicyIdMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*cachePolicyModel)
	if !ok {
		panic(fmt.Sprintf("id can only be called on a CachePolicy, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if model.CachePolicyId != "" {
		return model.CachePolicyId
	} else {
		return utils.DeferString(func() string {
			curr := s.GetCoinFrom(model.coin, []int{1, 3})
			if curr == nil {
				panic("could not find find/create version of " + model.coin.VarName().Id())
			}

			cp := curr.(*cachePolicyModel)
			if cp.CachePolicyId == "" {
				panic("cache policy id is still not set in " + cp.which)
			}
			return cp.CachePolicyId
		})
	}
}

var _ driverbottom.HasMethods = &cachePolicyModel{}

package cfront

import (
	"fmt"

	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type cachePolicyModel struct {
	loc           *errorsink.Location
	name          string
	minttl        driverbottom.Expr
	CachePolicyId string
}

func NewCachePolicyModel(loc *errorsink.Location, name string) *cachePolicyModel {
	return &cachePolicyModel{loc: loc, name: name}
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
			if model.CachePolicyId == "" {
				panic("id is still not set")
			}
			return model.CachePolicyId
		})
	}
}

var _ driverbottom.HasMethods = &cachePolicyModel{}

package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/utils"
)

type RoleModel struct {
	name string
	coin corebottom.CoinId

	assumption corebottom.PolicyActionList
	managed    []driverbottom.Expr
	inline     []corebottom.PolicyActionList
}

func (r *RoleModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "arn":
		return &deferredArnMethod{}
	default:
		log.Fatalf("there is no method %s on Role", name)
		return nil
	}
}

type deferredArnMethod struct {
}

func (rmm *deferredArnMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	e := on.Eval(s)
	var coin corebottom.CoinId
	switch model := e.(type) {
	case *RoleModel:
		coin = model.coin
	default:
		panic("invalid type")
	}
	return utils.DeferString(func() string {
		curr := s.GetCoinFrom(coin, []int{1, 3})
		if curr == nil {
			panic("could not find find/create version of " + coin.VarName().Id())
		}

		model := curr.(*RoleAWSModel)

		if model.role == nil {
			panic("vpc link is still not set")
		}
		return *model.role.Arn
	})
}

type RoleAWSModel struct {
	role     *types.Role
	policies []string
}

func (r *RoleAWSModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "arn":
		return &arnMethod{me: r.role}
	default:
		log.Fatalf("there is no method %s on Role", name)
		return nil
	}
}

type arnMethod struct {
	me *types.Role
}

func (a *arnMethod) Invoke(storage driverbottom.RuntimeStorage, obj driverbottom.Expr, args []driverbottom.Expr) any {
	if len(args) != 0 {
		return fmt.Errorf("arn method does not take arguments")
	}
	return *a.me.Arn
}

var _ driverbottom.Method = &arnMethod{}
var _ driverbottom.HasMethods = &RoleAWSModel{}

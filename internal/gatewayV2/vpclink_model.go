package gatewayV2

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type VPCLinkModel struct {
	name string
	loc  *errorsink.Location
	coin corebottom.CoinId

	subnets driverbottom.Expr
	groups  driverbottom.Expr
}

type VPCLinkAWSModel struct {
	link *types.VpcLink
	coin corebottom.CoinId
}

func (model *VPCLinkModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "id":
		return &deferredIdMethod{}
	case "subnets":
		return &subnetsMethod{}
	case "securityGroups":
		return &sgMethod{}
	}
	return nil
}

func (model *VPCLinkAWSModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "id":
		return &deferredIdMethod{}
	}
	return nil
}

type deferredIdMethod struct {
}

func (rmm *deferredIdMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	e := on.Eval(s)
	var coin corebottom.CoinId
	switch model := e.(type) {
	case *VPCLinkAWSModel:
		if model.link != nil {
			ret, ok := utils.AsStringer(model.link.VpcLinkId)
			if !ok {
				panic("not a string")
			}
			return ret
		}
		coin = model.coin
	case *VPCLinkModel:
		coin = model.coin
	default:
		panic("invalid type")
	}
	return utils.DeferString(func() string {
		curr := s.GetCoinFrom(coin, []int{1, 3})
		if curr == nil {
			panic("could not find find/create version of " + coin.VarName().Id())
		}

		link := curr.(*VPCLinkAWSModel)

		if link.link == nil {
			panic("vpc link is still not set")
		}
		return *link.link.VpcLinkId
	})
}

type subnetsMethod struct {
}

func (a *subnetsMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*VPCLinkModel)
	if !ok {
		panic(fmt.Sprintf("zoneId can only be called on a domain, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	return model.subnets
}

type sgMethod struct {
}

func (a *sgMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*VPCLinkModel)
	if !ok {
		panic(fmt.Sprintf("zoneId can only be called on a domain, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	return model.groups
}

var _ driverbottom.HasMethods = &VPCLinkModel{}
var _ driverbottom.HasMethods = &VPCLinkAWSModel{}

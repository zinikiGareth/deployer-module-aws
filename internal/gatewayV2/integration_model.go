package gatewayV2

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type IntegrationAWSModel struct {
	coin        corebottom.CoinId
	integration *types.Integration
}

func (i *IntegrationAWSModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "integrationId":
		return &intIdMethod{}
	}
	return nil
}

type intIdMethod struct {
}

func (a *intIdMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*IntegrationAWSModel)
	if !ok {
		panic(fmt.Sprintf("id can only be called on an integration, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if *model.integration.IntegrationId != "" {
		return *model.integration.IntegrationId
	} else {
		return getIntIdLater(s, model.coin)
	}
}

func getIntIdLater(s driverbottom.RuntimeStorage, coin corebottom.CoinId) fmt.Stringer {
	return utils.DeferString(func() string {
		curr := s.GetCoinFrom(coin, []int{1, 3})
		if curr == nil {
			panic("could not find find/create version of " + coin.VarName().Id())
		}

		currModel := curr.(*IntegrationAWSModel)
		if *currModel.integration.IntegrationId == "" {
			panic("api id is still not set")
		}
		return *currModel.integration.IntegrationId
	})

}

type IntegrationModel struct {
	name string
	loc  *errorsink.Location
	coin corebottom.CoinId

	api      fmt.Stringer
	region   fmt.Stringer
	itype    fmt.Stringer
	pfv      fmt.Stringer
	uri      driverbottom.Expr
	connType fmt.Stringer
	connId   fmt.Stringer
}

func (i *IntegrationModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "integrationId":
		return &intIdMethodD{}
	}
	return nil
}

type intIdMethodD struct {
}

func (a *intIdMethodD) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*IntegrationModel)
	if !ok {
		panic(fmt.Sprintf("id can only be called on an integration, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	return getIntIdLater(s, model.coin)
}

var _ driverbottom.HasMethods = &IntegrationAWSModel{}
var _ driverbottom.HasMethods = &IntegrationModel{}

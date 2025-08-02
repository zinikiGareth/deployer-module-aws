package gatewayV2

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type ApiAWSModel struct {
	coin corebottom.CoinId
	api  *types.Api
}

func (model *ApiAWSModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "id":
		return &idMethod{}
	}
	return nil
}

type idMethod struct {
}

func (a *idMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*ApiAWSModel)
	if !ok {
		panic(fmt.Sprintf("id can only be called on an api, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if *model.api.ApiId != "" {
		return *model.api.ApiId
	} else {
		return getIdLater(s, model.coin)
	}
}

func getIdLater(s driverbottom.RuntimeStorage, coin corebottom.CoinId) fmt.Stringer {
	return utils.DeferString(func() string {
		curr := s.GetCoinFrom(coin, []int{1, 3})
		if curr == nil {
			return "" // the api didn't exist
		}

		currModel := curr.(*ApiAWSModel)
		if *currModel.api.ApiId == "" {
			panic("api id is still not set")
		}
		return *currModel.api.ApiId
	})

}

type ApiModel struct {
	name     string
	loc      *errorsink.Location
	coin     corebottom.CoinId
	protocol types.ProtocolType
	rse      fmt.Stringer
}

func (model *ApiModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "id":
		return &idMethodD{}
	}
	return nil
}

type idMethodD struct {
}

func (a *idMethodD) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*ApiModel)
	if !ok {
		panic(fmt.Sprintf("id can only be called on an api, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	return getIdLater(s, model.coin)
}

var _ driverbottom.HasMethods = &ApiAWSModel{}
var _ driverbottom.HasMethods = &ApiModel{}

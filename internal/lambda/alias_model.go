package lambda

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/utils"
)

type LambdaAliasAWSModel struct {
	name string
	coin corebottom.CoinId

	function string
	arn      string
}

// this is only the bit after "arn:aws:apigateway:api-region:" which is common to all
// and the api-region is only known to the api
func (model *LambdaAliasAWSModel) makeIntegrationUri() string {
	return fmt.Sprintf("lambda:path/2015-03-31/functions/%s:%s/invocations", model.function, model.name)
}

func (model *LambdaAliasAWSModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "arn":
		return &aliasArnMethod{}
	case "integrationUri":
		return &aliasIntegrationUriMethod{}
	}
	return nil
}

type aliasArnMethod struct {
}

func (a *aliasArnMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*LambdaAliasAWSModel)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a lambda, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if model.arn != "" {
		return model.arn
	} else {
		return getAliasArnLater(s, model.coin)
	}
}

func getAliasArnLater(s driverbottom.RuntimeStorage, coin corebottom.CoinId) fmt.Stringer {
	return utils.DeferString(func() string {
		curr := s.GetCoinFrom(coin, []int{1, 3})
		if curr == nil {
			panic("could not find find/create version of " + coin.VarName().Id())
		}

		currModel := curr.(*LambdaAliasAWSModel)
		if currModel.arn == "" {
			panic("alias arn is still not set")
		}
		return currModel.arn
	})

}

type aliasIntegrationUriMethod struct {
}

func (a *aliasIntegrationUriMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*LambdaAliasAWSModel)
	if !ok {
		panic(fmt.Sprintf("integrationUri can only be called on a lambda alias, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if model.arn != "" {
		return model.makeIntegrationUri()
	} else {
		return utils.DeferString(func() string {
			curr := s.GetCoinFrom(model.coin, []int{1, 3})
			if curr == nil {
				panic("could not find find/create version of " + model.coin.VarName().Id())
			}

			currModel := curr.(*LambdaAliasAWSModel)
			if currModel.arn == "" {
				panic("lambda alias still not found")
			}
			return currModel.makeIntegrationUri()
		})
	}
}

var _ driverbottom.HasMethods = &LambdaModel{}

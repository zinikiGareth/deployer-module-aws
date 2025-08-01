package gatewayV2

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type DeploymentModel struct {
	name string
	loc  *errorsink.Location
	coin corebottom.CoinId

	api fmt.Stringer
}

/*
func (model *LambdaModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "arn":
		return &desiredArnMethod{}
	}
	return nil
}

type desiredArnMethod struct {
}

func (a *desiredArnMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*LambdaModel)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a lambda, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}

	// We are not going to find it here ... return a deferString
	return getArnLater(s, model.coin)
}
*/

type DeploymentAWSModel struct {
	name         string
	deploymentId string
}

/*
func (model *LambdaAWSModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "arn":
		return &arnMethod{}
			case "distributionId":
				return &distributionIdMethod{}
			case "domainName":
				return &domainNameMethod{}
			}
			return nil
		}

type arnMethod struct {
}

func (a *arnMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*LambdaAWSModel)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a lambda, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if *model.config.FunctionArn != "" {
		return *model.config.FunctionArn
	} else {
		return getArnLater(s, model.coin)
	}
}

func getArnLater(s driverbottom.RuntimeStorage, coin corebottom.CoinId) fmt.Stringer {
	return utils.DeferString(func() string {
		curr := s.GetCoinFrom(coin, []int{1, 3})
		if curr == nil {
			panic("could not find find/create version of " + coin.VarName().Id())
		}

		currModel := curr.(*LambdaAWSModel)
		if *currModel.config.FunctionArn == "" {
			panic("domain name is still not set")
		}
		return *currModel.config.FunctionArn
	})

}

var _ driverbottom.HasMethods = &LambdaModel{}
*/

package lambda

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/s3"
)

type LambdaModel struct {
	name string
	loc  *errorsink.Location
	coin corebottom.CoinId

	code    *s3.S3Location
	runtime driverbottom.Expr
	handler driverbottom.Expr
	role    driverbottom.Expr
}

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

type LambdaAWSModel struct {
	name string
	coin corebottom.CoinId

	config *types.FunctionConfiguration
}

// this is only the bit after "arn:aws:apigateway:api-region:" which is common to all
// and the api-region is only known to the api
func (model *LambdaAWSModel) makeIntegrationUri() string {
	return fmt.Sprintf("lambda:path//2015-03-31/functions/%s/invocations", *model.config.FunctionArn)
}

func (model *LambdaAWSModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "arn":
		return &arnMethod{}
	case "integrationUri":
		return &integrationUriMethod{}
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

type integrationUriMethod struct {
}

func (a *integrationUriMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*LambdaAWSModel)
	if !ok {
		panic(fmt.Sprintf("integrationUri can only be called on a lambda, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if model.config != nil {
		return model.makeIntegrationUri()
	} else {
		return utils.DeferString(func() string {
			curr := s.GetCoinFrom(model.coin, []int{1, 3})
			if curr == nil {
				panic("could not find find/create version of " + model.coin.VarName().Id())
			}

			currModel := curr.(*LambdaAWSModel)
			if currModel.config == nil {
				panic("lambda is still not found")
			}
			return currModel.makeIntegrationUri()
		})
	}
}

var _ driverbottom.HasMethods = &LambdaModel{}

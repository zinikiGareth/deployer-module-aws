package lambda

import (
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
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

	arn string
}

type LambdaAWSModel struct {
	name string
	// coin corebottom.CoinId

	found *lambda.GetFunctionOutput
	// created *lambda.CreateFunctionOutput
}

func (model *LambdaModel) ObtainMethod(name string) driverbottom.Method {
	/*
		switch name {
		case "arn":
			return &arnMethod{}
		case "distributionId":
			return &distributionIdMethod{}
		case "domainName":
			return &domainNameMethod{}
		}
	*/
	return nil
}

/*
type arnMethod struct {
}

	func (a *arnMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
		e := on.Eval(s)
		model, ok := e.(*DistributionModel)
		if !ok {
			panic(fmt.Sprintf("arn can only be called on a distribution, not a %T", e))
		}
		if len(args) != 0 {
			panic("invalid number of arguments")
		}
		if model.arn != "" {
			return model.arn
		} else {
			return utils.DeferString(func() string {
				curr := s.GetCoinFrom(model.coin, []int{1, 3})
				if curr == nil {
					panic("could not find find/create version of " + model.coin.VarName().Id())
				}

				currModel := curr.(*DistributionModel)
				if currModel.arn == "" {
					panic("domain name is still not set")
				}
				return currModel.arn
			})
		}
	}

type domainNameMethod struct {
}

	func (a *domainNameMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
		e := on.Eval(s)
		model, ok := e.(*DistributionModel)
		if !ok {
			panic(fmt.Sprintf("domainName can only be called on a distribution, not a %T", e))
		}
		if len(args) != 0 {
			panic("invalid number of arguments")
		}
		if model.domainName != "" {
			return model.domainName
		} else {
			return utils.DeferString(func() string {
				curr := s.GetCoinFrom(model.coin, []int{1, 3})
				if curr == nil {
					panic("could not find find/create version of " + model.coin.VarName().Id())
				}

				currModel := curr.(*DistributionModel)
				if currModel.domainName == "" {
					panic("domain name is still not set")
				}
				return currModel.domainName
			})
		}
	}

type distributionIdMethod struct {
}

	func (a *distributionIdMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
		e := on.Eval(s)
		model, ok := e.(*DistributionModel)
		if !ok {
			panic(fmt.Sprintf("distributionId can only be called on a distribution, not a %T", e))
		}
		if len(args) != 0 {
			panic("invalid number of arguments")
		}
		if model.distroId != "" {
			return model.distroId
		} else {
			return utils.DeferString(func() string {
				curr := s.GetCoinFrom(model.coin, []int{1, 3})
				if curr == nil {
					panic("could not find find/create version of " + model.coin.VarName().Id())
				}

				currModel := curr.(*DistributionModel)
				if currModel.distroId == "" {
					panic("domain name is still not set")
				}
				return currModel.distroId
			})
		}
	}
*/

var _ driverbottom.HasMethods = &LambdaModel{}

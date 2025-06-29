package acm

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type certificateModel struct {
	loc              *errorsink.Location
	name             string
	coin             corebottom.CoinId
	validationMethod fmt.Stringer
	hzid             string
	arn              string
	sans             []string
}

func (c *certificateModel) Loc() *errorsink.Location {
	return c.loc
}

func (c *certificateModel) ShortDescription() string {
	return fmt.Sprintf("acm.Certificate[%s]", c.name)
}

func (c *certificateModel) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

func (acmc *certificateModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "arn":
		return &arnMethod{}
	}
	return nil
}

type arnMethod struct {
}

func (a *arnMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	cc, ok := e.(*certificateModel)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a certificate, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if cc.arn != "" {
		return cc.arn
	} else {
		return utils.DeferString(func() string {
			curr := s.GetCoinFrom(cc.coin, []int{1, 3})
			if curr == nil {
				panic("could not find find/create version of " + cc.coin.VarName().Id())
			}

			cert := curr.(*certificateModel)
			if cert.arn == "" {
				panic("cert arn is still not set")
			}
			return cert.arn
		})
	}
}

func NewCertificateModel(loc *errorsink.Location, coin corebottom.CoinId) *certificateModel {
	return &certificateModel{loc: loc, coin: coin}
}

var _ driverbottom.Describable = &certificateModel{}
var _ driverbottom.HasMethods = &certificateModel{}

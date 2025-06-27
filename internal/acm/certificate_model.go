package acm

import (
	"fmt"

	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type certificateModel struct {
	loc              *errorsink.Location
	name             string
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
			return cc.arn
		})
	}
}

func NewCertificateModel(loc *errorsink.Location) *certificateModel {
	return &certificateModel{loc: loc}
}

var _ driverbottom.Describable = &certificateModel{}
var _ driverbottom.HasMethods = &certificateModel{}

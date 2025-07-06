package dynamodb

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type tableModel struct {
	loc              *errorsink.Location
	name             string
	coin             corebottom.CoinId
	validationMethod fmt.Stringer
	hzid             string
	arn              string
	sans             []string
}

func (c *tableModel) Loc() *errorsink.Location {
	return c.loc
}

func (c *tableModel) ShortDescription() string {
	return fmt.Sprintf("acm.Certificate[%s]", c.name)
}

func (c *tableModel) DumpTo(to driverbottom.IndentWriter) {
	to.Intro("Certificate[%s]", c.name)
	to.AttrsWhere(c)
	if c.validationMethod != nil {
		to.TextAttr("validationMethod", c.validationMethod.String())
	}
	to.TextAttr("hzid", c.hzid)
	to.TextAttr("arn", c.arn)
	for k, s := range c.sans {
		to.TextAttr(fmt.Sprintf("san%d", k), s)
	}
	to.EndAttrs()
}

func (acmc *tableModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	}
	return nil
}

func NewTableModel(loc *errorsink.Location, coin corebottom.CoinId) *tableModel {
	return &tableModel{loc: loc, coin: coin}
}

var _ driverbottom.Describable = &tableModel{}
var _ driverbottom.HasMethods = &tableModel{}

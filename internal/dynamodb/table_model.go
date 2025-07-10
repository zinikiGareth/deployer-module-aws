package dynamodb

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type tableModel struct {
	loc  *errorsink.Location
	name string
	coin corebottom.CoinId
	arn  string

	fields []*DynamoFieldExpr
}

func (c *tableModel) Loc() *errorsink.Location {
	return c.loc
}

func (c *tableModel) ShortDescription() string {
	return fmt.Sprintf("dynamo.Table[%s]", c.name)
}

func (c *tableModel) DumpTo(to driverbottom.IndentWriter) {
	to.Intro("dynamo.Table[%s]", c.name)
	to.AttrsWhere(c)
	to.TextAttr("arn", c.arn)
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

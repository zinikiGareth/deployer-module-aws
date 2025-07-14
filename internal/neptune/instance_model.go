package neptune

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type instanceModel struct {
	loc  *errorsink.Location
	name string
	coin corebottom.CoinId

	cluster     string
	instanceClz fmt.Stringer
	status      string
	arn         string
}

func (c *instanceModel) Loc() *errorsink.Location {
	return c.loc
}

func (c *instanceModel) ShortDescription() string {
	return fmt.Sprintf("acm.Certificate[%s]", c.name)
}

func (c *instanceModel) DumpTo(to driverbottom.IndentWriter) {
	to.Intro("Cluster[%s]", c.name)
	to.AttrsWhere(c)
	to.EndAttrs()
}

func (acmc *instanceModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	}
	return nil
}

func NewInstanceModel(loc *errorsink.Location, coin corebottom.CoinId, name string, arn string) *instanceModel {
	return &instanceModel{loc: loc, coin: coin, name: name, arn: arn}
}

var _ driverbottom.Describable = &instanceModel{}
var _ driverbottom.HasMethods = &instanceModel{}

package neptune

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type subnetModel struct {
	loc  *errorsink.Location
	name string
	coin corebottom.CoinId

	arn string
}

func (c *subnetModel) Loc() *errorsink.Location {
	return c.loc
}

func (c *subnetModel) ShortDescription() string {
	return fmt.Sprintf("neptune.SubnetGroup[%s]", c.name)
}

func (c *subnetModel) DumpTo(to driverbottom.IndentWriter) {
	to.Intro("SubnetGroup[%s]", c.name)
	to.AttrsWhere(c)
	to.EndAttrs()
}

func (acmc *subnetModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	}
	return nil
}

func NewSubnetGroupModel(loc *errorsink.Location, coin corebottom.CoinId, name string, arn string) *subnetModel {
	return &subnetModel{loc: loc, coin: coin, name: name, arn: arn}
}

var _ driverbottom.Describable = &subnetModel{}
var _ driverbottom.HasMethods = &subnetModel{}

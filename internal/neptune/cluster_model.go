package neptune

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type clusterModel struct {
	loc  *errorsink.Location
	name string
	coin corebottom.CoinId

	subnetGroup string
	minCapacity float64
	maxCapacity float64

	arn string
}

func (c *clusterModel) Loc() *errorsink.Location {
	return c.loc
}

func (c *clusterModel) ShortDescription() string {
	return fmt.Sprintf("acm.Certificate[%s]", c.name)
}

func (c *clusterModel) DumpTo(to driverbottom.IndentWriter) {
	to.Intro("Cluster[%s]", c.name)
	to.AttrsWhere(c)
	to.EndAttrs()
}

func (acmc *clusterModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	}
	return nil
}

func NewClusterModel(loc *errorsink.Location, coin corebottom.CoinId, name string, arn string) *clusterModel {
	return &clusterModel{loc: loc, coin: coin, name: name, arn: arn}
}

var _ driverbottom.Describable = &clusterModel{}
var _ driverbottom.HasMethods = &clusterModel{}

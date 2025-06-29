package cfront

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type cbModel struct {
	loc  *errorsink.Location
	name string

	pp             any
	rhp            any
	targetOriginId any
	cpId           fmt.Stringer
}

func (d *cbModel) Loc() *errorsink.Location {
	return d.loc
}

func (d *cbModel) ShortDescription() string {
	return "CacheBehaviour[" + d.name + "]"
}

func (d *cbModel) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("CacheBehavior %s", d.name)
	iw.AttrsWhere(d)
	// iw.NestedAttr("cpId", d.cpId)
	// iw.NestedAttr("pp", d.pp)
	// iw.NestedAttr("rhp", d.rhp)
	// iw.NestedAttr("toid", d.toid)
	iw.EndAttrs()
}

func (d *cbModel) Complete() types.CacheBehavior {
	toi := utils.AsString(d.targetOriginId)
	pp := utils.AsString(d.pp)
	rhp := utils.AsString(d.rhp)
	cpId := d.cpId.String()
	return types.CacheBehavior{TargetOriginId: &toi, PathPattern: &pp, ViewerProtocolPolicy: types.ViewerProtocolPolicyRedirectToHttps, CachePolicyId: &cpId, ResponseHeadersPolicyId: &rhp}
}

var _ driverbottom.Describable = &cbModel{}

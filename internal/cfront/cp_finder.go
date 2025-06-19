package cfront

import (
	"fmt"
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type CachePolicyFinder struct {
	tools *corebottom.Tools

	loc  *errorsink.Location
	name string
}

func (acm *CachePolicyFinder) Loc() *errorsink.Location {
	return acm.loc
}

func (acm *CachePolicyFinder) ShortDescription() string {
	return "aws.Cloudfront.CachePolicy[" + acm.name + "]"
}

func (acm *CachePolicyFinder) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.Cloudfront.CachePolicy[")
	iw.AttrsWhere(acm)
	iw.TextAttr("named", acm.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (cfc *CachePolicyFinder) BuildModel(pres driverbottom.ValuePresenter) {
	eq := cfc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	log.Printf("%v\n", awsEnv)
	pres.Present(cfc)

	panic("not implemented")
}

func (cfc *CachePolicyFinder) UpdateReality() {
}

func (cfc *CachePolicyFinder) String() string {
	return fmt.Sprintf("FindCloudFrontCachePolicy[%s]", cfc.name)
}

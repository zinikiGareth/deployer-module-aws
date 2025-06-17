package cfront

import (
	"fmt"
	"log"

	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type CachePolicyFinder struct {
	tools *pluggable.Tools

	loc  *errorsink.Location
	name string
}

func (acm *CachePolicyFinder) Loc() *errorsink.Location {
	return acm.loc
}

func (acm *CachePolicyFinder) ShortDescription() string {
	return "aws.Cloudfront.CachePolicy[" + acm.name + "]"
}

func (acm *CachePolicyFinder) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.Cloudfront.CachePolicy[")
	iw.AttrsWhere(acm)
	iw.TextAttr("named", acm.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (cfc *CachePolicyFinder) BuildModel(pres pluggable.ValuePresenter) {
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

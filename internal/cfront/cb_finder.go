package cfront

import (
	"fmt"
	"log"

	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type CacheBehaviorFinder struct {
	tools *external.Tools

	loc  *errorsink.Location
	name string
}

func (acm *CacheBehaviorFinder) Loc() *errorsink.Location {
	return acm.loc
}

func (acm *CacheBehaviorFinder) ShortDescription() string {
	return "aws.Cloudfront.CacheBehavior[" + acm.name + "]"
}

func (acm *CacheBehaviorFinder) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.Cloudfront.CacheBehavior[")
	iw.AttrsWhere(acm)
	iw.TextAttr("named", acm.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (cfc *CacheBehaviorFinder) BuildModel(pres pluggable.ValuePresenter) {
	eq := cfc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	log.Printf("%v\n", awsEnv)
	pres.Present(cfc)

	panic("not implemented")
}

func (cfc *CacheBehaviorFinder) UpdateReality() {
}

func (cfc *CacheBehaviorFinder) String() string {
	return fmt.Sprintf("FindCloudFrontCacheBehavior[%s]", cfc.name)
}

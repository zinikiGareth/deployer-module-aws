package cfront

import (
	"fmt"
	"log"

	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type RHPFinder struct {
	tools *external.Tools

	loc  *errorsink.Location
	name string
}

func (acm *RHPFinder) Loc() *errorsink.Location {
	return acm.loc
}

func (acm *RHPFinder) ShortDescription() string {
	return "aws.Cloudfront.RHP[" + acm.name + "]"
}

func (acm *RHPFinder) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.Cloudfront.RHP[")
	iw.AttrsWhere(acm)
	iw.TextAttr("named", acm.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (cfc *RHPFinder) BuildModel(pres pluggable.ValuePresenter) {
	eq := cfc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	log.Printf("%v\n", awsEnv)
	pres.Present(cfc)

	panic("not implemented")
}

func (cfc *RHPFinder) UpdateReality() {
}

func (cfc *RHPFinder) String() string {
	return fmt.Sprintf("FindCloudFrontRHP[%s]", cfc.name)
}

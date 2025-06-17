package cfront

import (
	"fmt"
	"log"

	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type OACFinder struct {
	tools *pluggable.Tools

	loc  *errorsink.Location
	name string
}

func (acm *OACFinder) Loc() *errorsink.Location {
	return acm.loc
}

func (acm *OACFinder) ShortDescription() string {
	return "aws.Cloudfront.OAC[" + acm.name + "]"
}

func (acm *OACFinder) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.Cloudfront.OAC[")
	iw.AttrsWhere(acm)
	iw.TextAttr("named", acm.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (cfc *OACFinder) BuildModel(pres pluggable.ValuePresenter) {
	eq := cfc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	log.Printf("%v\n", awsEnv)
	pres.Present(cfc)

	panic("not implemented")
}

func (cfc *OACFinder) UpdateReality() {
}

func (cfc *OACFinder) String() string {
	return fmt.Sprintf("FindCloudFrontOAC[%s]", cfc.name)
}

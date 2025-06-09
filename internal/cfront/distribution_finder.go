package cfront

import (
	"fmt"
	"log"

	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/deployer/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type distributionFinder struct {
	tools *pluggable.Tools

	loc  *errorsink.Location
	name string
}

func (acm *distributionFinder) Loc() *errorsink.Location {
	return acm.loc
}

func (acm *distributionFinder) ShortDescription() string {
	return "aws.CertificateManager.Certificate[" + acm.name + "]"
}

func (acm *distributionFinder) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.CertificateManager.Certificate[")
	iw.AttrsWhere(acm)
	iw.TextAttr("named", acm.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (cfc *distributionFinder) Prepare(pres pluggable.ValuePresenter) {
	eq := cfc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	log.Printf("%v\n", awsEnv)
	pres.Present(cfc)

	panic("not implemented")
}

func (cfc *distributionFinder) Execute() {
}

func (cfc *distributionFinder) String() string {
	return fmt.Sprintf("FindCloudFrontDistribution[%s]", cfc.name)
}

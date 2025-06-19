package acm

import (
	"fmt"
	"log"

	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/pluggable"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type certificateFinder struct {
	tools *external.Tools

	loc  *errorsink.Location
	name string
}

func (acm *certificateFinder) Loc() *errorsink.Location {
	return acm.loc
}

func (acm *certificateFinder) ShortDescription() string {
	return "aws.CertificateManager.Certificate[" + acm.name + "]"
}

func (acm *certificateFinder) DumpTo(iw pluggable.IndentWriter) {
	iw.Intro("aws.CertificateManager.Certificate[")
	iw.AttrsWhere(acm)
	iw.TextAttr("named", acm.name)
	iw.EndAttrs()
}

// This is called during the "Prepare" phase
func (acm *certificateFinder) BuildModel(pres pluggable.ValuePresenter) {
	eq := acm.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	log.Printf("%v\n", awsEnv)
	pres.Present(acm)

	panic("not implemented")
}

func (acm *certificateFinder) UpdateReality() {
}

func (acm *certificateFinder) String() string {
	return fmt.Sprintf("FindDomainName[%s]", acm.name)
}

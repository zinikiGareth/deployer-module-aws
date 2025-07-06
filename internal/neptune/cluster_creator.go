package neptune

import (
	"fmt"
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type clusterCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	teardown corebottom.TearDown
	props    map[driverbottom.Identifier]driverbottom.Expr
}

func (cc *clusterCreator) Loc() *errorsink.Location {
	return cc.loc
}

func (cc *clusterCreator) ShortDescription() string {
	return "aws.Neptune.Cluster[" + cc.name + "]"
}

func (cc *clusterCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.Neptune.Cluster[")
	iw.AttrsWhere(cc)
	iw.TextAttr("named", cc.name)
	iw.EndAttrs()
}

func (cc *clusterCreator) CoinId() corebottom.CoinId {
	return cc.coin
}

func (cc *clusterCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	log.Printf("want to find neptune cluster %s", cc.name)
	// ae := acmc.tools.Recall.ObtainDriver("aws.AwsEnv")
	// awsEnv, ok := ae.(*env.AwsEnv)
	// if !ok {
	// 	panic("could not cast env to AwsEnv")
	// }

	// if len(certs) == 0 {
	// 	log.Printf("there were no certs found for %s\n", acmc.name)
	// 	pres.NotFound()
	// } else {
	// 	model := NewCertificateModel(acmc.loc, acmc.coin)
	// 	model.name = acmc.name

	// 	// log.Printf("found %d certs for %s\n", len(certs), acmc.name)
	// 	model.arn = certs[0]

	// 	// acmc.describeCertificate(acmc.arn)
	// 	// acmc.tools.Storage.Bind(acmc.coin, model)
	// 	pres.Present(model)
	// }
}

func (cc *clusterCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	// model := NewCertificateModel(acmc.loc, acmc.coin)
	// for k, p := range acmc.props {
	// 	v := acmc.tools.Storage.Eval(p)
	// 	switch k.Id() {
	// 	case "Domain":
	// 		domain, ok := v.(myroute53.ExportedDomain)
	// 		if !ok {
	// 			log.Fatalf("Domain did not point to a domain instance")
	// 		}
	// 		model.hzid = domain.HostedZoneId()
	// 	case "SubjectAlternativeNames":
	// 		san, ok := utils.AsStringList(v)
	// 		if !ok {
	// 			justString, ok := v.(string)
	// 			if !ok {
	// 				log.Fatalf("SubjectAlternativeNames must be a list of strings")
	// 				return
	// 			} else {
	// 				san = []string{justString}
	// 			}
	// 		}
	// 		model.sans = san
	// 	case "ValidationMethod":
	// 		meth, ok := utils.AsStringer(v)
	// 		if !ok {
	// 			log.Fatalf("ValidationMethod must be a string")
	// 			return
	// 		}
	// 		model.validationMethod = meth
	// 	default:
	// 		log.Fatalf("certificate coin does not support a parameter %s\n", k.Id())
	// 	}
	// }
	// // acmc.tools.Storage.Bind(acmc.coin, model)
	// pres.Present(model)
}

func (cc *clusterCreator) UpdateReality() {
	// tmp := acmc.tools.Storage.GetCoin(acmc.coin, corebottom.DETERMINE_INITIAL_MODE)

	// if tmp != nil {
	// 	found := tmp.(*clusterModel)
	// 	log.Printf("certificate %s already existed for %s\n", found.arn, found.name)
	// 	acmc.tools.Storage.Adopt(acmc.coin, found)
	// 	return
	// }

	// desired := acmc.tools.Storage.GetCoin(acmc.coin, corebottom.DETERMINE_DESIRED_MODE).(*clusterModel)

	// created := NewCertificateModel(desired.loc, acmc.coin)
	// created.name = desired.name
	// created.hzid = desired.hzid

	// vm := types.ValidationMethod(desired.validationMethod.String())
	// if vm == "" {
	// 	vm = types.ValidationMethodDns
	// }

	// input := acm.RequestCertificateInput{DomainName: &acmc.name, ValidationMethod: vm}
	// if len(desired.sans) > 0 {
	// 	input.SubjectAlternativeNames = desired.sans
	// }
	// req, err := acmc.client.RequestCertificate(context.TODO(), &input)
	// if err != nil {
	// 	log.Printf("failed to request cert %s: %v\n", acmc.name, err)
	// }
	// log.Printf("requested cert for %s: %s\n", acmc.name, *req.CertificateArn)
	// created.arn = *req.CertificateArn

	// // Check if we still need to validate it ...

	// // acmc.describeCertificate(*req.CertificateArn)

	// utils.ExponentialBackoff(func() bool {
	// 	return acmc.tryToValidateCert(created.arn, created.hzid)
	// })

	// acmc.tools.Storage.Bind(acmc.coin, created)
}

func (cc *clusterCreator) TearDown() {
	// tmp := acmc.tools.Storage.GetCoin(acmc.coin, corebottom.DETERMINE_INITIAL_MODE)

	// if tmp == nil {
	// 	log.Printf("no certificate existed for %s\n", acmc.name)
	// 	return
	// }

	// found := tmp.(*clusterModel)
	// log.Printf("you have asked to tear down certificate for %s (arn: %s) with mode %s\n", found.name, found.arn, acmc.teardown.Mode())
	// switch acmc.teardown.Mode() {
	// case "preserve":
	// 	log.Printf("not deleting certificate %s because teardown mode is 'preserve'", found.name)
	// case "delete":
	// 	log.Printf("deleting certificate for %s with teardown mode 'delete'", found.name)
	// 	DeleteCertificate(acmc.client, found.arn)
	// default:
	// 	log.Printf("cannot handle teardown mode '%s' for bucket %s", acmc.teardown.Mode(), found.name)
	// }
}

func (cc *clusterCreator) String() string {
	return fmt.Sprintf("ClusterCreator[%s]", cc.name)
}

var _ corebottom.Ensurable = &clusterCreator{}

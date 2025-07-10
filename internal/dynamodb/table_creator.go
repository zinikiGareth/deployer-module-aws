package dynamodb

import (
	"context"
	"fmt"
	"log"

	ht "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type tableCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	teardown corebottom.TearDown
	props    map[driverbottom.Identifier]driverbottom.Expr

	client *dynamodb.Client
}

func (tc *tableCreator) Loc() *errorsink.Location {
	return tc.loc
}

func (tc *tableCreator) ShortDescription() string {
	return "aws.CertificateManager.Certificate[" + tc.name + "]"
}

func (tc *tableCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.CertificateManager.Certificate[")
	iw.AttrsWhere(tc)
	iw.TextAttr("named", tc.name)
	iw.EndAttrs()
}

func (tc *tableCreator) CoinId() corebottom.CoinId {
	return tc.coin
}

func (tc *tableCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	log.Printf("want to find dynamo table %s", tc.name)

	ae := tc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := ae.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}
	tc.client = awsEnv.DynamoClient()

	table := tc.findTableCalled(tc.name)
	if table == nil {
		log.Printf("no dynamo table found called %s\n", tc.name)
		pres.NotFound()
	} else {
		// 	model := NewCertificateModel(tc.loc, tc.coin)
		// 	model.name = tc.name

		// 	// log.Printf("found %d certs for %s\n", len(certs), tc.name)
		// 	model.arn = certs[0]

		// 	// tc.describeCertificate(tc.arn)
		// 	// tc.tools.Storage.Bind(tc.coin, model)
		pres.Present(table)
	}
}

func (tc *tableCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	model := NewTableModel(tc.loc, tc.coin)
	model.name = tc.name
	for k, p := range tc.props {
		v := tc.tools.Storage.Eval(p)
		switch k.Id() {
		case "Fields":
			list, ok := v.([]any)
			if !ok {
				panic("Fields was not a list")
			}
			for _, le := range list {
				dfe, ok := le.(*DynamoFieldExpr)
				if ok {
					model.attrs = append(model.attrs, tc.makeDRF(dfe))
				} else {
					panic("field was not a *DynamoFieldExpr")
				}
			}
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
		}
	}
	pres.Present(model)
}

func (tc *tableCreator) UpdateReality() {
	tmp := tc.tools.Storage.GetCoin(tc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*tableModel)
		log.Printf("table %s already existed for %s\n", found.arn, found.name)
		tc.tools.Storage.Adopt(tc.coin, found)
		return
	}

	desired := tc.tools.Storage.GetCoin(tc.coin, corebottom.DETERMINE_DESIRED_MODE).(*tableModel)

	created := NewTableModel(desired.loc, tc.coin)
	created.name = desired.name
	// created.hzid = desired.hzid

	// vm := types.ValidationMethod(desired.validationMethod.String())
	// if vm == "" {
	// 	vm = types.ValidationMethodDns
	// }

	input := dynamodb.CreateTableInput{TableName: &created.name, AttributeDefinitions: desired.attrs}
	// if len(desired.sans) > 0 {
	// 	input.SubjectAlternativeNames = desired.sans
	// }
	table, err := tc.client.CreateTable(context.TODO(), &input)
	if err != nil {
		log.Printf("failed to create table %s: %v\n", tc.name, err)
		panic("table creation failed")
	}
	log.Printf("asked to create table for %s: %s\n", tc.name, *table.TableDescription.TableArn)
	created.arn = *table.TableDescription.TableArn

	utils.ExponentialBackoff(func() bool {
		return tc.waitForTable(tc.name)
	})

	log.Printf("created table %s for %s", created.arn, created.name)
	tc.tools.Storage.Bind(tc.coin, created)
}

func (tc *tableCreator) TearDown() {
	// tmp := tc.tools.Storage.GetCoin(tc.coin, corebottom.DETERMINE_INITIAL_MODE)

	// if tmp == nil {
	// 	log.Printf("no certificate existed for %s\n", tc.name)
	// 	return
	// }

	// found := tmp.(*certificateModel)
	// log.Printf("you have asked to tear down certificate for %s (arn: %s) with mode %s\n", found.name, found.arn, tc.teardown.Mode())
	// switch tc.teardown.Mode() {
	// case "preserve":
	// 	log.Printf("not deleting certificate %s because teardown mode is 'preserve'", found.name)
	// case "delete":
	// 	log.Printf("deleting certificate for %s with teardown mode 'delete'", found.name)
	// 	DeleteCertificate(tc.client, found.arn)
	// default:
	// 	log.Printf("cannot handle teardown mode '%s' for bucket %s", tc.teardown.Mode(), found.name)
	// }
}

func (tc *tableCreator) findTableCalled(name string) *tableModel {
	table, err := tc.client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{TableName: &name})
	if !tableExists(err) {
		return nil
	}
	log.Fatalf("table = %T %v\n", table, table)
	return nil
}

func (tc *tableCreator) waitForTable(name string) bool {
	_, err := tc.client.DescribeTable(context.TODO(), &dynamodb.DescribeTableInput{TableName: &name})
	return tableExists(err)
}

func tableExists(err error) bool {
	if err == nil {
		return true
	}
	e1, ok := err.(*smithy.OperationError)
	if ok {
		e2, ok := e1.Err.(*ht.ResponseError)
		if ok {
			if e2.ResponseError.Response.StatusCode == 400 {
				switch e4 := e2.Err.(type) {
				case *types.ResourceNotFoundException:
					return false
				default:
					log.Printf("error: %T %v", e4, e4)
					panic("what error?")
				}
			}
			log.Fatalf("error: %T %v %T %v", e2.Response.Status, e2.Response.Status, e2.ResponseError.Response.StatusCode, e2.ResponseError.Response.StatusCode)
		}
		log.Fatalf("error: %T %v", e1.Err, e1.Err)
	}
	log.Fatalf("getting clusters failed: %T %v", err, err)
	panic("failed")
}

func (tc *tableCreator) String() string {
	return fmt.Sprintf("DynamoTable[%s]", tc.name)
}

func (tc *tableCreator) makeDRF(dfe *DynamoFieldExpr) types.AttributeDefinition {
	return types.AttributeDefinition{AttributeName: &dfe.name, AttributeType: tc.figureType(dfe.loc, dfe.ftype)}
}

func (tc *tableCreator) figureType(loc *errorsink.Location, ty string) types.ScalarAttributeType {
	switch ty {
	case "string":
		return types.ScalarAttributeTypeS
	case "number":
		return types.ScalarAttributeTypeN
	case "bool":
		return types.ScalarAttributeTypeB
	default:
		tc.tools.Reporter.ReportAtf(loc, "invalid dynamo field type: %s", ty)
		return types.ScalarAttributeTypeS
	}
}

var _ corebottom.Ensurable = &tableCreator{}

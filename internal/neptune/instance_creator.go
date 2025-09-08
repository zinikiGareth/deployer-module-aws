package neptune

import (
	"context"
	"fmt"
	"log"
	"strings"

	ht "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type instanceCreator struct {
	tools *corebottom.Tools

	loc  *errorsink.Location
	name string
	coin corebottom.CoinId
	// todo: allow @teardown finalSnapshot
	// will require @finalShapshotIdentifier
	teardown corebottom.TearDown
	props    map[driverbottom.Identifier]driverbottom.Expr

	client *neptune.Client
}

func (cc *instanceCreator) Loc() *errorsink.Location {
	return cc.loc
}

func (cc *instanceCreator) ShortDescription() string {
	return "aws.Neptune.Instance[" + cc.name + "]"
}

func (cc *instanceCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.Neptune.Instance[")
	iw.AttrsWhere(cc)
	iw.TextAttr("named", cc.name)
	iw.EndAttrs()
}

func (cc *instanceCreator) CoinId() corebottom.CoinId {
	return cc.coin
}

func (cc *instanceCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	log.Printf("want to find neptune instance %s", cc.name)
	ae := cc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := ae.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	cc.client = awsEnv.NeptuneClient()

	model := cc.findInstancesNamed(cc.name)

	if model == nil {
		log.Printf("instance %s not found\n", cc.name)
		pres.NotFound()
	} else {
		log.Printf("instance found for %s\n", cc.name)
		pres.Present(model)
	}
}

func (cc *instanceCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	model := NewInstanceModel(cc.loc, cc.coin, cc.name, "")
	for k, p := range cc.props {
		v := cc.tools.Storage.Eval(p)
		switch k.Id() {
		case "Cluster":
			cluster, ok := v.(*clusterModel)
			if !ok {
				log.Fatalf("Cluster did not point to a cluster model")
			}
			model.cluster = cluster.name
		case "InstanceClass":
			clz, ok := utils.AsStringer(v)
			if !ok {
				log.Fatalf("ValidationMethod must be a string")
				return
			}
			model.instanceClz = clz
		default:
			log.Fatalf("neptune instance does not support a parameter %s\n", k.Id())
		}
	}
	// if model.subnetGroup == "" {
	// 	cc.tools.Reporter.ReportAtf(cc.loc, "SubnetGroupName required for Neptune Instance")
	// 	return
	// }
	log.Printf("have desired neptune config for %s\n", model.name)
	pres.Present(model)
}

func (cc *instanceCreator) UpdateReality() {
	tmp := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*instanceModel)
		log.Printf("instance %s already existed for %s\n", found.arn, found.name)
		if found.status != "available" {
			utils.ExponentialBackoff(func() bool {
				return cc.waitForCreation(found)
			})
		}
		cc.tools.Storage.Adopt(cc.coin, found)
		return
	}

	desired := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_DESIRED_MODE).(*instanceModel)

	created := NewInstanceModel(desired.loc, cc.coin, cc.name, "")

	neptuneName := "neptune" // because of some way that AWS centralizes DB creation
	instClz := desired.instanceClz.String()
	if !strings.HasPrefix(instClz, "db.") {
		instClz = "db." + instClz
	}
	create, err := cc.client.CreateDBInstance(context.TODO(), &neptune.CreateDBInstanceInput{DBInstanceIdentifier: &cc.name, Engine: &neptuneName, DBClusterIdentifier: &desired.cluster, DBInstanceClass: &instClz})
	if err != nil {
		log.Fatalf("failed to create instance %s: %v\n", cc.name, err)
	}
	created.arn = *create.DBInstance.DBInstanceArn
	log.Printf("initiated request to create instance %s: %s %s\n", cc.name, *create.DBInstance.DBInstanceStatus, *create.DBInstance.DBInstanceArn)

	utils.ExponentialBackoff(func() bool {
		return cc.waitForCreation(created)
	})

	log.Printf("created neptune instance %s %s", created.name, created.arn)
	cc.tools.Storage.Bind(cc.coin, created)
}

func (cc *instanceCreator) TearDown() {
	tmp := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp == nil {
		log.Printf("no instance existed for %s\n", cc.name)
		return
	}

	found := tmp.(*instanceModel)
	log.Printf("you have asked to tear down neptune instance for %s (arn: %s) with mode %s\n", found.name, found.arn, cc.teardown.Mode())
	// todo: allow @teardown finalSnapshot
	// will require @finalShapshotIdentifier
	switch cc.teardown.Mode() {
	case "preserve":
		log.Printf("not deleting instance %s because teardown mode is 'preserve'", found.name)
	case "delete":
		log.Printf("deleting instance for %s with teardown mode 'delete'", found.name)
		cc.deleteInstance(found, "")
	default:
		log.Printf("cannot handle teardown mode '%s' for bucket %s", cc.teardown.Mode(), found.name)
	}

	utils.ExponentialBackoff(func() bool {
		return cc.waitForDeletion(found)
	})

	log.Printf("deleted neptune instance %s", found.name)
}

func (cc *instanceCreator) findInstancesNamed(name string) *instanceModel {
	instances, err := cc.client.DescribeDBInstances(context.TODO(), &neptune.DescribeDBInstancesInput{DBInstanceIdentifier: &name})
	if !instanceExists(err) {
		return nil
	}
	if len(instances.DBInstances) == 0 {
		return nil
	} else if len(instances.DBInstances) > 1 {
		log.Fatalf("More than one instance called %s found", name)
		panic("multiple instances")
	} else {
		c1 := instances.DBInstances[0]
		ret := NewInstanceModel(cc.loc, cc.coin, *c1.DBInstanceIdentifier, *c1.DBInstanceArn)
		ret.status = *instances.DBInstances[0].DBInstanceStatus
		return ret
	}
}

func instanceExists(err error) bool {
	if err == nil {
		return true
	}
	e1, ok := err.(*smithy.OperationError)
	if ok {
		e2, ok := e1.Err.(*ht.ResponseError)
		if ok {
			if e2.ResponseError.Response.StatusCode == 404 {
				return false
			}
			log.Fatalf("error: %T %v %T %v", e2.Response.Status, e2.Response.Status, e2.ResponseError.Response.StatusCode, e2.ResponseError.Response.StatusCode)
		}
		log.Fatalf("error: %T %v", e1.Err, e1.Err)
	}
	log.Fatalf("getting clusters failed: %T %v", err, err)
	panic("failed")
}

func (cc *instanceCreator) deleteInstance(instance *instanceModel, finalSnapshotId string) {
	args := &neptune.DeleteDBInstanceInput{DBInstanceIdentifier: &instance.name}
	if finalSnapshotId != "" {
		args.FinalDBSnapshotIdentifier = &finalSnapshotId
	} else {
		skip := true
		args.SkipFinalSnapshot = &skip
	}
	_, err := cc.client.DeleteDBInstance(context.TODO(), args)
	if err != nil {
		log.Fatalf("deleting instance failed: %T %v", err, err)
	}
}

func (cc *instanceCreator) waitForCreation(instance *instanceModel) bool {
	instances, err := cc.client.DescribeDBInstances(context.TODO(), &neptune.DescribeDBInstancesInput{DBInstanceIdentifier: &instance.name})
	if !instanceExists(err) || len(instances.DBInstances) == 0 {
		log.Printf("no instances found with name %s\n", instance.name)
		return false
	}
	c := instances.DBInstances[0]
	if c.DBInstanceStatus != nil {
		if *c.DBInstanceStatus == "available" {
			return true
		} else {
			log.Printf("status was %s, not available", *c.DBInstanceStatus)
		}
	} else {
		log.Printf("status was nil")
	}
	return false
}

func (cc *instanceCreator) waitForDeletion(instance *instanceModel) bool {
	instances, err := cc.client.DescribeDBInstances(context.TODO(), &neptune.DescribeDBInstancesInput{DBInstanceIdentifier: &instance.name})
	if !instanceExists(err) || len(instances.DBInstances) == 0 {
		return true
	}
	/*
		c := instances.DBInstances[0]
			if c.Status != nil {
				if *c.Status == "deleting" || *c.Status == "available" {
					log.Printf("instance %s still exists with status %s\n", instance.name, *c.Status)
				} else {
					log.Fatalf("status was %s, not available or deleting", *c.Status)
				}
			} else {
				log.Printf("instance %s still exists with nil status\n", instance.name)
			}
	*/
	return false
}

func (cc *instanceCreator) String() string {
	return fmt.Sprintf("instanceCreator[%s]", cc.name)
}

var _ corebottom.Ensurable = &instanceCreator{}

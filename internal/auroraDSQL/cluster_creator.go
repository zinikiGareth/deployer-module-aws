package auroraDSQL

import (
	"context"
	"log"
	"reflect"

	ht "github.com/aws/aws-sdk-go-v2/aws/transport/http"

	"github.com/aws/aws-sdk-go-v2/service/dsql"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"

	"ziniki.org/deployer/coremod/pkg/corepkg"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type ClusterCreator struct {
	Client *dsql.Client
}

/*
	func (cc *ClusterCreator) ShortDescription() string {
		return "aws.Neptune.Cluster[" + cc.name + "]"
	}

	func (cc *ClusterCreator) DumpTo(iw driverbottom.IndentWriter) {
		iw.Intro("aws.Neptune.Cluster[")
		iw.AttrsWhere(cc)
		iw.TextAttr("named", cc.name)
		iw.EndAttrs()
	}
*/

func (cc *ClusterCreator) DetermineInitialState(creator *corepkg.CoreCreator, pres corebottom.ValuePresenter) {
	creator.GetEnv("aws.AwsEnv", reflect.TypeFor[*env.AwsEnv](), "AuroraClient", "Client")
	log.Printf("client = %T %p\n", cc.Client, cc.Client)
	model := cc.findClustersNamed(creator.Name())
	if model == nil {
		log.Printf("cluster %s not found\n", creator.Name())
		pres.NotFound()
	} else {
		log.Printf("cluster found for %s\n", creator.Name())
		pres.Present(model)
	}
}

func (cc *ClusterCreator) DetermineDesiredState(creator *corepkg.CoreCreator, pres corebottom.ValuePresenter) {
	/*
		model := NewClusterModel(cc.loc, cc.coin, cc.name, "")
		for k, p := range cc.props {
			v := cc.tools.Storage.Eval(p)
			switch k.Id() {
			case "SubnetGroupName":
				subnetGroup, ok := v.(*subnetModel)
				if !ok {
					log.Fatalf("SubnetGroupName did not point to a subnet model")
				}
				model.subnetGroup = subnetGroup.name
			case "MinCapacity":
				cap, ok := v.(float64)
				if ok {
					model.minCapacity = cap
				}
			case "MaxCapacity":
				cap, ok := v.(float64)
				if ok {
					model.maxCapacity = cap
				}
			default:
				log.Fatalf("neptune cluster does not support a parameter %s\n", k.Id())
			}
		}
		if model.subnetGroup == "" {
			cc.tools.Reporter.ReportAtf(cc.loc, "SubnetGroupName required for Neptune Cluster")
			return
		}
		if model.maxCapacity != 0 && model.minCapacity != 0 {
			// that's fine
		} else if model.maxCapacity != 0 || model.minCapacity != 0 {
			cc.tools.Reporter.ReportAtf(cc.loc, "must either specify neither MinCapacity or MaxCapacity or both")
			return
		}
		log.Printf("have desired neptune config for %s\n", model.name)
		pres.Present(model)

	*/
}

func (cc *ClusterCreator) UpdateReality(creator *corepkg.CoreCreator, initial any, desired any) {
	/*
		tmp := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_INITIAL_MODE)

		if tmp != nil {
			found := tmp.(*clusterModel)
			log.Printf("cluster %s already existed for %s\n", found.arn, found.name)
			cc.tools.Storage.Adopt(cc.coin, found)
			return
		}

		desired := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_DESIRED_MODE).(*clusterModel)

		created := NewClusterModel(desired.loc, cc.coin, cc.name, "")

		neptuneName := "neptune" // because of some way that AWS centralizes DB creation
		ci := &neptune.CreateDBClusterInput{DBClusterIdentifier: &cc.name, Engine: &neptuneName, DBSubnetGroupName: &desired.subnetGroup}
		minCap := 1.0
		maxCap := 1.0
		if desired.minCapacity != 0 && desired.maxCapacity != 0 {
			scaling := &types.ServerlessV2ScalingConfiguration{MinCapacity: &minCap, MaxCapacity: &maxCap}
			ci.ServerlessV2ScalingConfiguration = scaling
		}
		create, err := cc.client.CreateDBCluster(context.TODO(), ci)
		if err != nil {
			log.Fatalf("failed to create cluster %s: %v\n", cc.name, err)
		}
		created.arn = *create.DBCluster.DBClusterArn
		log.Printf("initiated request to create cluster %s: %s %s\n", cc.name, *create.DBCluster.Status, *create.DBCluster.DBClusterArn)

		utils.ExponentialBackoff(func() bool {
			return cc.waitForCreation(created)
		})

		log.Printf("created neptune cluster %s %s", created.name, created.arn)
		cc.tools.Storage.Bind(cc.coin, created)

	*/
}

func (cc *ClusterCreator) TearDown(creator *corepkg.CoreCreator, initial any) {
	/*
		tmp := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_INITIAL_MODE)

		if tmp == nil {
			log.Printf("no cluster existed for %s\n", cc.name)
			return
		}

		found := tmp.(*clusterModel)
		log.Printf("you have asked to tear down neptune cluster for %s (arn: %s) with mode %s\n", found.name, found.arn, cc.teardown.Mode())
		// todo: allow @teardown finalSnapshot
		// will require @finalShapshotIdentifier
		switch cc.teardown.Mode() {
		case "preserve":
			log.Printf("not deleting cluster %s because teardown mode is 'preserve'", found.name)
		case "delete":
			log.Printf("deleting cluster for %s with teardown mode 'delete'", found.name)
			cc.deleteCluster(found, "")
		default:
			log.Printf("cannot handle teardown mode '%s' for bucket %s", cc.teardown.Mode(), found.name)
		}

		utils.ExponentialBackoff(func() bool {
			return cc.waitForDeletion(found)
		})

		log.Printf("deleted neptune cluster %s", found.name)
	*/
}

type clusterModel struct {
	arn string
	id  string
}

func (cc *ClusterCreator) findClustersNamed(name string) *clusterModel {
	var tok *string
	for {
		clusters, err := cc.Client.ListClusters(context.TODO(), &dsql.ListClustersInput{NextToken: tok})
		if !clusterExists(err) {
			return nil
		}
		for _, c := range clusters.Clusters {
			if *c.Identifier == name {
				return &clusterModel{arn: *c.Arn, id: *c.Identifier}
			}
		}
		if clusters.NextToken == nil {
			return nil
		}
		tok = clusters.NextToken
	}
}

func clusterExists(err error) bool {
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

/*
func (cc *ClusterCreator) deleteCluster(cluster *clusterModel, finalSnapshotId string) {
	args := &neptune.DeleteDBClusterInput{DBClusterIdentifier: &cluster.name}
	if finalSnapshotId != "" {
		args.FinalDBSnapshotIdentifier = &finalSnapshotId
	} else {
		skip := true
		args.SkipFinalSnapshot = &skip
	}
	_, err := cc.client.DeleteDBCluster(context.TODO(), args)
	if err != nil {
		log.Fatalf("deleting cluster failed: %T %v", err, err)
	}
}

func (cc *ClusterCreator) waitForCreation(cluster *clusterModel) bool {
	clusters, err := cc.client.DescribeDBClusters(context.TODO(), &neptune.DescribeDBClustersInput{DBClusterIdentifier: &cluster.name})
	if !clusterExists(err) || len(clusters.DBClusters) == 0 {
		log.Printf("no clusters found with name %s\n", cluster.name)
		return false
	}
	c := clusters.DBClusters[0]
	if c.Status != nil {
		if *c.Status == "available" {
			return true
		} else {
			log.Printf("status was %s, not available", *c.Status)
		}
	} else {
		log.Printf("status was nil")
	}
	return false
}

func (cc *ClusterCreator) waitForDeletion(cluster *clusterModel) bool {
	clusters, err := cc.client.DescribeDBClusters(context.TODO(), &neptune.DescribeDBClustersInput{DBClusterIdentifier: &cluster.name})
	if !clusterExists(err) || len(clusters.DBClusters) == 0 {
		return true
	}
	c := clusters.DBClusters[0]
	if c.Status != nil {
		if *c.Status == "deleting" || *c.Status == "available" {
			log.Printf("cluster %s still exists with status %s\n", cluster.name, *c.Status)
		} else {
			log.Fatalf("status was %s, not available or deleting", *c.Status)
		}
	} else {
		log.Printf("cluster %s still exists with nil status\n", cluster.name)
	}
	return false
}

func (cc *ClusterCreator) String() string {
	return fmt.Sprintf("ClusterCreator[%s]", cc.name)
}

*/

var _ corepkg.CreationStrategy = &ClusterCreator{}

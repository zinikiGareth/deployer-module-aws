package auroraDSQL

import (
	"context"
	"log"
	"reflect"

	ht "github.com/aws/aws-sdk-go-v2/aws/transport/http"

	"github.com/aws/aws-sdk-go-v2/service/dsql"
	"github.com/aws/aws-sdk-go-v2/service/dsql/types"
	"github.com/aws/smithy-go"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/utils"

	"ziniki.org/deployer/coremod/pkg/corepkg"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type ClusterCreator struct {
	Client *dsql.Client
}

func (cc *ClusterCreator) DetermineInitialState(creator *corepkg.CoreCreator, pres corebottom.ValuePresenter) {
	creator.GetEnv("aws.AwsEnv", reflect.TypeFor[*env.AwsEnv](), "AuroraClient", "Client")
	log.Printf("client = %T %p\n", cc.Client, cc.Client)
	model := cc.findClusterNamed(creator.Name())
	if model == nil {
		log.Printf("cluster %s not found\n", creator.Name())
		pres.NotFound()
	} else {
		log.Printf("cluster found for %s\n", creator.Name())
		pres.Present(model)
	}
}

func (cc *ClusterCreator) DetermineDesiredState(creator *corepkg.CoreCreator, pres corebottom.ValuePresenter) {
	model := &clusterModel{core: creator}
	log.Printf("have desired aurora config for %s\n", creator.Name())
	pres.Present(model)
}

func (cc *ClusterCreator) UpdateReality(creator *corepkg.CoreCreator, initial any, desired any) {
	if initial != nil {
		found := initial.(*clusterAWSModel)
		log.Printf("cluster %s already existed for %s\n", found.arn, creator.Name())
		creator.Adopt(found)
		return
	}

	tags := map[string]string{}
	tags["Name"] = creator.Name()
	ci := &dsql.CreateClusterInput{Tags: tags}

	create, err := cc.Client.CreateCluster(context.TODO(), ci)
	if err != nil {
		log.Fatalf("failed to create cluster %s: %v\n", creator.Name(), err)
	}
	created := &clusterAWSModel{id: *create.Identifier, arn: *create.Arn}

	log.Printf("initiated request to create cluster %s: %v %s\n", creator.Name(), create.Status, *create.Arn)
	utils.ExponentialBackoff(func() bool {
		return cc.waitForCreation(created)
	})

	log.Printf("created aurora dsql cluster %s %s", creator.Name(), created.arn)
	creator.Created(created)
}

func (cc *ClusterCreator) TearDown(creator *corepkg.CoreCreator, initial any, teardown corebottom.TearDown) {
	if initial == nil {
		log.Printf("no cluster existed for %s\n", creator.Name())
		return
	}

	found := initial.(*clusterAWSModel)
	log.Printf("you have asked to tear down aurora dsql cluster for %s (arn: %s) with mode %s\n", creator.Name(), found.arn, teardown.Mode())
	// todo: allow @teardown finalSnapshot
	// will require @finalShapshotIdentifier
	switch teardown.Mode() {
	case "preserve":
		log.Printf("not deleting cluster %s because teardown mode is 'preserve'", creator.Name())
	case "delete":
		log.Printf("deleting cluster for %s with teardown mode 'delete'", creator.Name())
		cc.deleteCluster(creator, found)
	default:
		log.Printf("cannot handle teardown mode '%s' for bucket %s", teardown.Mode(), creator.Name())
	}

	utils.ExponentialBackoff(func() bool {
		return cc.waitForDeletion(creator, found)
	})

	log.Printf("deleted neptune cluster %s", creator.Name())
}

type clusterModel struct {
	core *corepkg.CoreCreator
}

func (c *clusterModel) ObtainMethod(name string) driverbottom.Method {
	return c.core.DeferredMethod(name)
}

type clusterAWSModel struct {
	arn string
	id  string
}

// ObtainMethod implements driverbottom.HasMethods.
func (c *clusterAWSModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "arn":
		return corepkg.SimpleMethod(func(storage driverbottom.RuntimeStorage, obj driverbottom.Expr) any {
			eval := storage.Eval(obj).(*clusterAWSModel)
			return eval.arn
		})
	case "id":
		return corepkg.SimpleMethod(func(storage driverbottom.RuntimeStorage, obj driverbottom.Expr) any {
			eval := storage.Eval(obj).(*clusterAWSModel)
			return eval.id
		})
	default:
		panic("no such method " + name)
	}
}

func (cc *ClusterCreator) findClusterNamed(name string) *clusterAWSModel {
	var tok *string
	for {
		clusters, err := cc.Client.ListClusters(context.TODO(), &dsql.ListClustersInput{NextToken: tok})
		if err != nil {
			panic(err)
		}
		for _, c := range clusters.Clusters {
			tags, err := cc.Client.ListTagsForResource(context.TODO(), &dsql.ListTagsForResourceInput{ResourceArn: c.Arn})
			if err != nil {
				panic(err)
			}
			if tags.Tags["Name"] == name {
				return &clusterAWSModel{arn: *c.Arn, id: *c.Identifier}
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
			log.Fatalf("error: %v %v", e2.ResponseError.Response.StatusCode, e2.Response.Status)
		}
		log.Fatalf("error: %T %v", e1.Err, e1.Err)
	}
	log.Fatalf("getting clusters failed: %T %v", err, err)
	panic("failed")
}

func (cc *ClusterCreator) waitForCreation(cluster *clusterAWSModel) bool {
	c, err := cc.Client.GetCluster(context.TODO(), &dsql.GetClusterInput{Identifier: &cluster.id})
	if !clusterExists(err) {
		log.Printf("no clusters found with id %s\n", cluster.id)
		return false
	}
	if c.Status == types.ClusterStatusActive {
		return true
	} else {
		log.Printf("status was %s, not available", c.Status)
	}
	return false
}

func (cc *ClusterCreator) deleteCluster(creator *corepkg.CoreCreator, cluster *clusterAWSModel) {
	f := false
	mod := &dsql.UpdateClusterInput{Identifier: &cluster.id, DeletionProtectionEnabled: &f}
	_, err := cc.Client.UpdateCluster(context.TODO(), mod)
	if err != nil {
		log.Fatalf("failed to remove deletion protection from cluster %s: %v", creator.Name(), err)
	}

	args := &dsql.DeleteClusterInput{Identifier: &cluster.id}
	_, err = cc.Client.DeleteCluster(context.TODO(), args)
	if err != nil {
		log.Fatalf("deleting cluster failed: %T %v", err, err)
	}
}

func (cc *ClusterCreator) waitForDeletion(creator *corepkg.CoreCreator, cluster *clusterAWSModel) bool {
	c, err := cc.Client.GetCluster(context.TODO(), &dsql.GetClusterInput{Identifier: &cluster.id})
	if !clusterExists(err) || c.Status == types.ClusterStatusDeleted {
		return true
	}
	if c.Status == types.ClusterStatusDeleting || c.Status == types.ClusterStatusActive {
		log.Printf("cluster %s still exists with status %v\n", creator.Name(), c.Status)
	} else {
		log.Fatalf("status was %v, not available or deleting", c.Status)
	}
	return false
}

var _ corepkg.CreationStrategy = &ClusterCreator{}
var _ driverbottom.HasMethods = &clusterModel{}
var _ driverbottom.HasMethods = &clusterAWSModel{}

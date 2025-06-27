package s3

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/policyjson"
)

type bucketModel struct {
	loc     *errorsink.Location
	storage driverbottom.RuntimeStorage
	id      corebottom.CoinId
	// May I say how much I hate that this is here, but we need it for Attach ...
	client *s3.Client

	name   string
	region fmt.Stringer

	policy string
}

func (b *bucketModel) Attach(doc corebottom.PolicyDocument) {
	// TODO: I assume we either have to merge or do duplicate detection
	policyJson, err := policyjson.BuildFrom(b.name, doc)
	if err != nil {
		log.Fatalf("could not build policy: %v", err)
	}
	newbm := &bucketModel{loc: b.loc, storage: b.storage, id: b.id, name: b.name, client: b.client, policy: policyJson}
	b.storage.Bind(b.id, newbm)
	b.client.PutBucketPolicy(context.TODO(), &s3.PutBucketPolicyInput{Bucket: &b.name, Policy: &policyJson})
	log.Printf("attached policy to bucket %s\n", b.name)
}

func (b *bucketModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "allResources":
		return &allResourcesMethod{}
	case "dnsName":
		return &dnsNameMethod{}
	}
	return nil
}

func (b *bucketModel) ObtainDest() corebottom.FileDest {
	return NewBucketTransfer(b.client, b.name)
}

type allResourcesMethod struct {
}

func (a *allResourcesMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	bucket, ok := e.(*bucketModel)
	if !ok {
		panic(fmt.Sprintf("allResources can only be called on a bucket, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	return fmt.Sprintf("arn:aws:s3:::%s/*", bucket.name)
}

// return something like "news.consolidator.info.s3.us-east-1.amazonaws.com"
type dnsNameMethod struct {
}

func (a *dnsNameMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	bucket, ok := e.(*bucketModel)
	if !ok {
		panic(fmt.Sprintf("domainName can only be called on a bucket, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	return fmt.Sprintf("%s.s3.%s.amazonaws.com", bucket.name, bucket.region)
}

var _ driverbottom.HasMethods = &bucketModel{}
var _ driverbottom.Method = &allResourcesMethod{}
var _ driverbottom.Method = &dnsNameMethod{}
var _ corebottom.PolicyAttacher = &bucketModel{}
var _ corebottom.DestHolder = &bucketModel{}

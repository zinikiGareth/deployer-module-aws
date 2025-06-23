package route53

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	r53types "github.com/aws/aws-sdk-go-v2/service/route53/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type cnameCreator struct {
	tools *corebottom.Tools

	loc      *errorsink.Location
	name     string
	pointsTo driverbottom.Expr
	zone     driverbottom.Expr
	teardown corebottom.TearDown

	client        *route53.Client
	alreadyExists bool
	otherDomain   any
	zoneId        string
}

func (p *cnameCreator) Loc() *errorsink.Location {
	return p.loc
}

func (p *cnameCreator) ShortDescription() string {
	return "aws.IAM.Policy[" + p.name + "]"
}

func (p *cnameCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.IAM.Policy[")
	iw.AttrsWhere(p)
	iw.TextAttr("named", p.name)
	iw.EndAttrs()
}

func (cc *cnameCreator) BuildModel(pres driverbottom.ValuePresenter) {
	eq := cc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	// cc.domainsClient = awsEnv.Route53DomainsClient()
	cc.client = awsEnv.Route53Client()

	fred, ok := cc.tools.Storage.EvalAsStringer(cc.zone)
	if !ok {
		panic("hello, world")
	}
	cc.zoneId = fred.String()
	pt := cc.pointsTo.Eval(cc.tools.Storage)
	cc.otherDomain = pt

	log.Printf("scanning zone %s\n", cc.zoneId)
	rrs, err := cc.client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{HostedZoneId: &cc.zoneId})
	if err != nil {
		panic(err)
	}

	for _, r := range rrs.ResourceRecordSets {
		// log.Printf("found rrs %s %s", *r.Name, *r.ResourceRecords[0].Value)
		if r.Type == "CNAME" && *r.Name == cc.name+"." {
			// TODO: we should also handle the case where it has changed
			log.Printf("already have %s %v\n", *r.Name, *r.ResourceRecords[0].Value)
			cc.alreadyExists = true
		}
	}
}

func (cc *cnameCreator) UpdateReality() {
	if cc.alreadyExists {
		log.Printf("CNAME %s already exists\n", cc.name)
		return
	}

	log.Printf("creating CNAME %s\n", cc.name)

	var ttl int64 = 300
	od, ok := cc.otherDomain.(string)
	if !ok {
		str, ok := cc.otherDomain.(fmt.Stringer)
		if !ok {
			panic("not a string or Stringer)")
		}
		od = str.String()
	}

	changes := r53types.ResourceRecordSet{Name: &cc.name, Type: "CNAME", TTL: &ttl, ResourceRecords: []r53types.ResourceRecord{{Value: &od}}}
	cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "CREATE", ResourceRecordSet: &changes}}}
	_, err := cc.client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &cc.zoneId, ChangeBatch: &cb})
	if err != nil {
		panic(err)
	}
}

func (cc *cnameCreator) TearDown() {
	if !cc.alreadyExists {
		log.Printf("CNAME %s already deleted\n", cc.name)
		return
	}
	log.Printf("need to remove a CNAME record for %s\n", cc.name)
	od, ok := cc.otherDomain.(string)
	if !ok {
		str, ok := cc.otherDomain.(fmt.Stringer)
		if !ok {
			panic("not a string or Stringer)")
		}
		od = str.String()
	}
	var ttl int64 = 300
	changes := r53types.ResourceRecordSet{Name: &cc.name, Type: "CNAME", TTL: &ttl, ResourceRecords: []r53types.ResourceRecord{{Value: &od}}}
	cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "DELETE", ResourceRecordSet: &changes}}}
	_, err := cc.client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &cc.zoneId, ChangeBatch: &cb})
	if err != nil {
		panic(err)
	}
}

func (p *cnameCreator) String() string {
	return fmt.Sprintf("EnsurePolicy[%s:%s]", "" /* eb.env.Region */, p.name)
}

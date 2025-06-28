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

type aliasCreator struct {
	tools *corebottom.Tools

	loc        *errorsink.Location
	name       string
	coin       corebottom.CoinId
	pointsTo   driverbottom.Expr
	updateZone driverbottom.Expr
	aliasZone  driverbottom.Expr
	teardown   corebottom.TearDown

	client        *route53.Client
	alreadyExists bool
	otherDomain   any
	updateZoneId  string
	aliasZoneId   string
}

func (p *aliasCreator) Loc() *errorsink.Location {
	return p.loc
}

func (p *aliasCreator) ShortDescription() string {
	return "aws.IAM.Policy[" + p.name + "]"
}

func (p *aliasCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.IAM.Policy[")
	iw.AttrsWhere(p)
	iw.TextAttr("named", p.name)
	iw.EndAttrs()
}

func (acmc *aliasCreator) CoinId() corebottom.CoinId {
	return acmc.coin
}

func (cc *aliasCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
}

func (cc *aliasCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	eq := cc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	// cc.domainsClient = awsEnv.Route53DomainsClient()
	cc.client = awsEnv.Route53Client()

	fred, ok := cc.tools.Storage.EvalAsStringer(cc.updateZone)
	if !ok {
		panic("hello, world")
	}
	cc.updateZoneId = fred.String()

	fred, ok = cc.tools.Storage.EvalAsStringer(cc.aliasZone)
	if !ok {
		panic("hello, world")
	}
	cc.aliasZoneId = fred.String()

	pt := cc.pointsTo.Eval(cc.tools.Storage)
	cc.otherDomain = pt

	log.Printf("scanning zone %s\n", cc.updateZoneId)
	rrs, err := cc.client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{HostedZoneId: &cc.updateZoneId})
	if err != nil {
		panic(err)
	}

	for _, r := range rrs.ResourceRecordSets {
		if r.AliasTarget == nil {
			continue
		}
		if r.Type == "A" && *r.Name == cc.name+"." {
			// TODO: we should also handle the case where it has changed
			log.Printf("already have A %s %v\n", *r.Name, *r.AliasTarget.DNSName)
			cc.alreadyExists = true
		}
	}
}

func (cc *aliasCreator) UpdateReality() {
	if cc.alreadyExists {
		log.Printf("alias %s already exists\n", cc.name)
		return
	}

	log.Printf("creating alias %s\n", cc.name)

	od, ok := cc.otherDomain.(string)
	if !ok {
		str, ok := cc.otherDomain.(fmt.Stringer)
		if !ok {
			panic("not a string or Stringer)")
		}
		od = str.String()
	}

	changes := r53types.ResourceRecordSet{Name: &cc.name, Type: "A", AliasTarget: &r53types.AliasTarget{DNSName: &od, HostedZoneId: &cc.aliasZoneId}}
	cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "CREATE", ResourceRecordSet: &changes}}}
	_, err := cc.client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &cc.updateZoneId, ChangeBatch: &cb})
	if err != nil {
		panic(err)
	}
}

func (cc *aliasCreator) TearDown() {
	if !cc.alreadyExists {
		log.Printf("alias %s already deleted\n", cc.name)
		return
	}
	log.Printf("need to remove an alias record for %s\n", cc.name)
	od, ok := cc.otherDomain.(string)
	if !ok {
		str, ok := cc.otherDomain.(fmt.Stringer)
		if !ok {
			panic("not a string or Stringer)")
		}
		od = str.String()
	}
	changes := r53types.ResourceRecordSet{Name: &cc.name, Type: "A", AliasTarget: &r53types.AliasTarget{DNSName: &od, HostedZoneId: &cc.aliasZoneId}}
	cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "DELETE", ResourceRecordSet: &changes}}}
	_, err := cc.client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &cc.updateZoneId, ChangeBatch: &cb})
	if err != nil {
		panic(err)
	}
}

func (p *aliasCreator) String() string {
	return fmt.Sprintf("EnsureAlias[%s:%s]", "" /* eb.env.Region */, p.name)
}

var _ corebottom.Ensurable = &aliasCreator{}

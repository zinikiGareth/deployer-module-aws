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

	loc      *errorsink.Location
	name     string
	coin     corebottom.CoinId
	props    map[driverbottom.Identifier]driverbottom.Expr
	teardown corebottom.TearDown

	client *route53.Client
}

func (ac *aliasCreator) Loc() *errorsink.Location {
	return ac.loc
}

func (ac *aliasCreator) ShortDescription() string {
	return "aws.Route53.Alias[" + ac.name + "]"
}

func (ac *aliasCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.Route53.ALIAS")
	iw.AttrsWhere(ac)
	iw.TextAttr("named", ac.name)
	iw.EndAttrs()
}

func (ac *aliasCreator) CoinId() corebottom.CoinId {
	return ac.coin
}

func (ac *aliasCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	var updZone driverbottom.Expr
	var aliasZone driverbottom.Expr
	seenErr := false
	for p, v := range ac.props {
		switch p.Id() {
		case "PointsTo":
		case "AliasZone":
			aliasZone = v

		case "UpdateZone":
			updZone = v
		default:
			ac.tools.Reporter.ReportAtf(p.Loc(), "invalid property for IAM policy: %s", p.Id())
			seenErr = true
		}
	}
	if !seenErr && updZone == nil {
		ac.tools.Reporter.ReportAtf(ac.loc, "no UpdateZone property was specified for %s", ac.name)
	}
	if !seenErr && aliasZone == nil {
		ac.tools.Reporter.ReportAtf(ac.loc, "no AliasZone property was specified for %s", ac.name)
	}

	updZoneId, ok := ac.tools.Storage.EvalAsStringer(updZone)
	if !ok {
		panic("hello, world")
	}

	aliasZoneId, ok := ac.tools.Storage.EvalAsStringer(aliasZone)
	if !ok {
		panic("hello, world")
	}

	eq := ac.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	// cc.domainsClient = awsEnv.Route53DomainsClient()
	ac.client = awsEnv.Route53Client()

	uz := updZoneId.String()
	log.Printf("scanning zone %s\n", uz)
	rrs, err := ac.client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{HostedZoneId: &uz})
	if err != nil {
		panic(err)
	}

	for _, r := range rrs.ResourceRecordSets {
		if r.AliasTarget == nil {
			continue
		}
		if r.Type == "A" && *r.Name == ac.name+"." {
			// TODO: we should also handle the case where it has changed
			log.Printf("already have A %s %v\n", *r.Name, *r.AliasTarget.DNSName)
			model := &aliasModel{loc: ac.loc, name: ac.name, otherDomain: *r.AliasTarget.DNSName, aliasZoneId: aliasZoneId.String(), updateZoneId: updZoneId.String()}
			pres.Present(model)
			return
		}
	}

	pres.NotFound()
}

func (ac *aliasCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var pointsTo driverbottom.Expr
	var updZone driverbottom.Expr
	var aliasZone driverbottom.Expr
	seenErr := false
	for p, v := range ac.props {
		switch p.Id() {
		case "PointsTo":
			pointsTo = v
		case "UpdateZone":
			updZone = v
		case "AliasZone":
			aliasZone = v
		default:
			ac.tools.Reporter.ReportAtf(p.Loc(), "invalid property for IAM policy: %s", p.Id())
			seenErr = true
		}
	}
	if !seenErr && updZone == nil {
		ac.tools.Reporter.ReportAtf(ac.loc, "no UpdateZone property was specified for %s", ac.name)
	}
	if !seenErr && aliasZone == nil {
		ac.tools.Reporter.ReportAtf(ac.loc, "no AliasZone property was specified for %s", ac.name)
	}
	if !seenErr && pointsTo == nil {
		ac.tools.Reporter.ReportAtf(ac.loc, "no PointsTo property was specified for %s", ac.name)
	}

	updZoneId, ok := ac.tools.Storage.EvalAsStringer(updZone)
	if !ok {
		panic("hello, world")
	}

	aliasZoneId, ok := ac.tools.Storage.EvalAsStringer(aliasZone)
	if !ok {
		panic("hello, world")
	}

	pt := pointsTo.Eval(ac.tools.Storage)

	model := &aliasModel{loc: ac.loc, name: ac.name, otherDomain: pt, updateZoneId: updZoneId.String(), aliasZoneId: aliasZoneId.String()}
	pres.Present(model)
}

func (ac *aliasCreator) UpdateReality() {
	tmp := ac.tools.Storage.GetCoin(ac.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*aliasModel)
		log.Printf("alias %s already exists\n", found.name)
		return
	}

	log.Printf("creating alias %s\n", ac.name)
	desired := ac.tools.Storage.GetCoin(ac.coin, corebottom.DETERMINE_DESIRED_MODE).(*aliasModel)

	created := &aliasModel{name: ac.name, loc: ac.loc}

	od, ok := desired.otherDomain.(string)
	if !ok {
		str, ok := desired.otherDomain.(fmt.Stringer)
		if !ok {
			panic("not a string or Stringer)")
		}
		od = str.String()
	}

	changes := r53types.ResourceRecordSet{Name: &ac.name, Type: "A", AliasTarget: &r53types.AliasTarget{DNSName: &od, HostedZoneId: &desired.aliasZoneId}}
	cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "CREATE", ResourceRecordSet: &changes}}}
	_, err := ac.client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &desired.updateZoneId, ChangeBatch: &cb})
	if err != nil {
		panic(err)
	}

	ac.tools.Storage.Bind(ac.coin, created)
}

func (ac *aliasCreator) TearDown() {
	tmp := ac.tools.Storage.GetCoin(ac.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp == nil {
		log.Printf("alias %s already deleted\n", ac.name)
		return
	}

	found := tmp.(*aliasModel)
	log.Printf("need to remove an alias record for %s\n", ac.name)
	od, ok := found.otherDomain.(string)
	if !ok {
		str, ok := found.otherDomain.(fmt.Stringer)
		if !ok {
			panic("not a string or Stringer)")
		}
		od = str.String()
	}
	changes := r53types.ResourceRecordSet{Name: &ac.name, Type: "A", AliasTarget: &r53types.AliasTarget{DNSName: &od, HostedZoneId: &found.aliasZoneId}}
	cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "DELETE", ResourceRecordSet: &changes}}}
	_, err := ac.client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &found.updateZoneId, ChangeBatch: &cb})
	if err != nil {
		panic(err)
	}
}

func (ac *aliasCreator) String() string {
	return fmt.Sprintf("EnsureAlias[%s:%s]", "" /* eb.env.Region */, ac.name)
}

var _ corebottom.Ensurable = &aliasCreator{}

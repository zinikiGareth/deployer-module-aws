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

	loc   *errorsink.Location
	name  string
	coin  corebottom.CoinId
	props map[driverbottom.Identifier]driverbottom.Expr

	client *route53.Client
}

func (cc *cnameCreator) Loc() *errorsink.Location {
	return cc.loc
}

func (cc *cnameCreator) ShortDescription() string {
	return "aws.Route53.CNAME[" + cc.name + "]"
}

func (cc *cnameCreator) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.Route53.CNAME")
	iw.AttrsWhere(cc)
	iw.TextAttr("named", cc.name)
	iw.EndAttrs()
}

func (cc *cnameCreator) CoinId() corebottom.CoinId {
	return cc.coin
}

func (cc *cnameCreator) DetermineInitialState(pres corebottom.ValuePresenter) {
	var zone driverbottom.Expr
	seenErr := false
	for p, v := range cc.props {
		switch p.Id() {
		case "PointsTo":
		case "Zone":
			zone = v
		default:
			cc.tools.Reporter.ReportAtf(cc.loc, "invalid property for IAM policy: %s", p.Id())
		}
	}
	if !seenErr && zone == nil {
		cc.tools.Reporter.ReportAtf(cc.loc, "no Zone property was specified for %s", cc.name)
	}
	eq := cc.tools.Recall.ObtainDriver("aws.AwsEnv")
	awsEnv, ok := eq.(*env.AwsEnv)
	if !ok {
		panic("could not cast env to AwsEnv")
	}

	cc.client = awsEnv.Route53Client()
	fred, ok := cc.tools.Storage.EvalAsStringer(zone)
	if !ok {
		panic("hello, world")
	}

	z := fred.String()
	log.Printf("scanning zone %s\n", z)
	rrs, err := cc.client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{HostedZoneId: &z})
	if err != nil {
		panic(err)
	}

	for _, r := range rrs.ResourceRecordSets {
		// log.Printf("found rrs %s %s", *r.Name, *r.ResourceRecords[0].Value)
		if r.Type == "CNAME" && *r.Name == cc.name+"." {
			// TODO: we should also handle the case where it has changed
			log.Printf("already have %s %v\n", *r.Name, *r.ResourceRecords[0].Value)
			model := &cnameModel{loc: cc.loc, name: cc.name, pointsTo: *r.ResourceRecords[0].Value, updateZoneId: z}
			pres.Present(model)
			return
		}
	}

	pres.NotFound()
}

func (cc *cnameCreator) DetermineDesiredState(pres corebottom.ValuePresenter) {
	var pointsTo driverbottom.Expr
	var zone driverbottom.Expr
	seenErr := false
	for p, v := range cc.props {
		switch p.Id() {
		case "PointsTo":
			pointsTo = v
		case "Zone":
			zone = v
		default:
			cc.tools.Reporter.ReportAtf(cc.loc, "invalid property for IAM policy: %s", p.Id())
			seenErr = true
		}
	}
	if !seenErr && zone == nil {
		cc.tools.Reporter.ReportAtf(cc.loc, "no Zone property was specified for %s", cc.name)
	}
	if !seenErr && pointsTo == nil {
		cc.tools.Reporter.ReportAtf(cc.loc, "no PointsTo property was specified for %s", cc.name)
	}

	zoneId, ok := cc.tools.Storage.EvalAsStringer(zone)
	if !ok {
		panic("hello, world")
	}
	pt := pointsTo.Eval(cc.tools.Storage)

	model := &cnameModel{loc: cc.loc, name: cc.name, pointsTo: pt, updateZoneId: zoneId.String()}
	pres.Present(model)

}

func (cc *cnameCreator) UpdateReality() {
	tmp := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp != nil {
		found := tmp.(*cnameModel)
		log.Printf("CNAME %s already exists\n", found.name)
		return
	}

	log.Printf("creating CNAME %s\n", cc.name)
	desired := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_DESIRED_MODE).(*cnameModel)

	created := &cnameModel{name: cc.name, loc: cc.loc}

	var ttl int64 = 300
	od, ok := desired.pointsTo.(string)
	if !ok {
		str, ok := desired.pointsTo.(fmt.Stringer)
		if !ok {
			log.Printf("pointsto was %T %p", desired.pointsTo, desired.pointsTo)
			panic("not a string or Stringer")
		}
		od = str.String()
	}

	changes := r53types.ResourceRecordSet{Name: &cc.name, Type: "CNAME", TTL: &ttl, ResourceRecords: []r53types.ResourceRecord{{Value: &od}}}
	cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "CREATE", ResourceRecordSet: &changes}}}
	_, err := cc.client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &desired.updateZoneId, ChangeBatch: &cb})
	if err != nil {
		panic(err)
	}

	cc.tools.Storage.Bind(cc.coin, created)
}

func (cc *cnameCreator) TearDown() {
	tmp := cc.tools.Storage.GetCoin(cc.coin, corebottom.DETERMINE_INITIAL_MODE)

	if tmp == nil {
		log.Printf("CNAME %s already deleted\n", cc.name)
		return
	}

	found := tmp.(*cnameModel)
	log.Printf("need to remove a CNAME record for %s\n", cc.name)
	od, ok := found.pointsTo.(string)
	if !ok {
		str, ok := found.pointsTo.(fmt.Stringer)
		if !ok {
			log.Printf("pointsto was %T %p", found.pointsTo, found.pointsTo)
			panic("not a string or Stringer")
		}
		od = str.String()
	}
	var ttl int64 = 300
	changes := r53types.ResourceRecordSet{Name: &cc.name, Type: "CNAME", TTL: &ttl, ResourceRecords: []r53types.ResourceRecord{{Value: &od}}}
	cb := r53types.ChangeBatch{Changes: []r53types.Change{{Action: "DELETE", ResourceRecordSet: &changes}}}
	_, err := cc.client.ChangeResourceRecordSets(context.TODO(), &route53.ChangeResourceRecordSetsInput{HostedZoneId: &found.updateZoneId, ChangeBatch: &cb})
	if err != nil {
		panic(err)
	}
}

func (cc *cnameCreator) String() string {
	return fmt.Sprintf("EnsureCNAME[%s:%s]", "" /* eb.env.Region */, cc.name)
}

var _ corebottom.Ensurable = &cnameCreator{}

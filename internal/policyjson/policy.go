package policyjson

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/external"
)

type policyJson struct {
	Version   string
	Statement []stmtJson
}

func makePolicyJson(name string, policy external.PolicyDocument) *policyJson {
	ret := &policyJson{}
	ret.Version = "2012-10-17"
	ret.Statement = []stmtJson{}
	for k, item := range policy.Items() {
		ret.Statement = append(ret.Statement, makeStmtJson(name, k, item))
	}
	return ret
}

type stmtJson struct {
	Sid       string
	Effect    string
	Action    any // can be string or []string
	Resource  any // can be string or []string
	Principal any // can be string or []string
}

func makeStmtJson(policyName string, k int, item external.PolicyEffect) stmtJson {
	ret := stmtJson{Sid: fmt.Sprintf("%s-sid-%d", policyName, k), Effect: item.Effect()}
	as := item.Actions()
	if len(as) == 0 {
		ret.Action = "*"
	} else if len(as) == 1 {
		ret.Action = as[0]
	} else {
		ret.Action = as
	}
	rs := item.Resources()
	if len(rs) == 0 {
		ret.Resource = "*"
	} else if len(rs) == 1 {
		ret.Resource = rs[0]
	} else {
		ret.Resource = rs
	}
	ps := item.Principals()
	if len(ps) == 0 {
		ret.Principal = "*"
	} else if len(ps) == 1 {
		ret.Principal = ps[0]
	} else {
		ret.Principal = ps
	}
	return ret
}

package policyjson

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/external"
)

type policyJson struct {
	Version   string
	Statement []stmtJson
}

func makePolicyJson(policy external.PolicyDocument) *policyJson {
	ret := &policyJson{}
	ret.Version = "2012-10-17"
	ret.Statement = []stmtJson{}
	for k, item := range policy.Items() {
		ret.Statement = append(ret.Statement, makeStmtJson(policy.Name(), k, item))
	}
	return ret
}

type stmtJson struct {
	Sid      string
	Effect   string
	Action   any // can be string or []string
	Resource any // can be string or []string
}

func makeStmtJson(policyName string, k int, item external.PolicyEffect) stmtJson {
	ret := stmtJson{Sid: fmt.Sprintf("%s-%d", policyName, k), Effect: item.Effect()}
	as := item.Actions()
	if len(as) == 0 {
		ret.Action = "*"
	}
	rs := item.Resources()
	if len(rs) == 0 {
		ret.Resource = "*"
	}
	return ret
}

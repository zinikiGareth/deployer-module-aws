package policyjson

import (
	"fmt"
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
)

type policyJson struct {
	Version   string
	Statement []stmtJson
}

func makePolicyJson(name string, policy corebottom.PolicyDocument, rules PolicyRules) *policyJson {
	ret := &policyJson{}
	ret.Version = "2012-10-17"
	ret.Statement = []stmtJson{}
	for k, item := range policy.Items() {
		ret.Statement = append(ret.Statement, makeStmtJson(name, k, item, rules))
	}
	return ret
}

type stmtJson struct {
	Sid       string
	Effect    string
	Action    any // can be string or []string
	Resource  any `json:",omitempty"` // can be string or []string
	Principal any `json:",omitempty"`
	Condition any `json:",omitempty"`
}

func makeStmtJson(policyName string, k int, item corebottom.PolicyEffect, rules PolicyRules) stmtJson {
	ret := stmtJson{Sid: fmt.Sprintf("%sSid%d", policyName, k), Effect: item.Effect()}
	as := item.Actions()
	if len(as) == 0 {
		ret.Action = "*"
	} else if len(as) == 1 {
		ret.Action = as[0]
	} else {
		ret.Action = as
	}
	rs := item.Resources()
	if rules.AllowResources() {
		if len(rs) == 0 {
			ret.Resource = "*"
		} else if len(rs) == 1 {
			ret.Resource = rs[0]
		} else {
			ret.Resource = rs
		}
	} else if len(rs) != 0 {
		panic("resources not allowed here")
	}
	ps := item.Principals()
	if len(ps) == 0 {
		ret.Principal = nil
	} else if len(ps) == 1 {
		ret.Principal = makePrincipal(ps[0])
	} else {
		rps := []map[string]string{}
		for _, p := range ps {
			rps = append(rps, makePrincipal(p))
		}
		ret.Principal = rps
	}

	conds := item.More()["Condition"]
	if len(conds) > 0 {
		if len(conds) == 1 {
			ret.Condition = makeCondition(conds[0])
		} else {
			cs := []map[string]any{}
			for _, c := range cs {
				cs = append(cs, makeCondition(c))
			}
			ret.Condition = cs
		}
	}
	return ret
}

func makePrincipal(p corebottom.PolicyPrincipal) map[string]string {
	ret := map[string]string{}
	ret[p.Key()] = p.Value()
	return ret
}

func makeCondition(p any) map[string]any {
	input, ok := p.(map[string]any)
	if !ok {
		return nil
	}
	ret := map[string]any{}
	for k, v := range input {
		mv, ok := v.(map[string]any)
		if !ok {
			continue
		}
		rm := map[string]string{}
		ret[k] = rm
		for l, r := range mv {
			switch rs := r.(type) {
			case string:
				rm[l] = rs
			case fmt.Stringer:
				rm[l] = rs.String()
			default:
				log.Fatalf("can't handle %T %v", r, r)
			}
		}
	}
	return ret
}

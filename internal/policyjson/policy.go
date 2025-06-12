package policyjson

import "ziniki.org/deployer/coremod/pkg/external"

type policyJson struct {
	Version   string
	Statement []stmtJson
}

func makePolicyJson(policy external.PolicyDocument) *policyJson {
	ret := policyJson{}
	ret.Version = "2012-10-17"
	ret.Statement = []stmtJson{}
	return &ret
}

type stmtJson struct {
}

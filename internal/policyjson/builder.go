package policyjson

import (
	"encoding/json"

	"ziniki.org/deployer/coremod/pkg/corebottom"
)

func BuildFrom(name string, policy corebottom.PolicyDocument, rules PolicyRules) (string, error) {
	p := makePolicyJson(name, policy, rules)

	bs, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

type PolicyRules interface {
	AllowResources() bool
}

type assemblyRules struct {
	allowResources bool
}

func (r *assemblyRules) AllowResources() bool {
	return r.allowResources
}

func StandardRules() PolicyRules {
	return &assemblyRules{allowResources: true}
}

func AssumeRoleRules() PolicyRules {
	return &assemblyRules{allowResources: false}
}

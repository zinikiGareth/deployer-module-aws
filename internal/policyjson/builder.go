package policyjson

import (
	"encoding/json"

	"ziniki.org/deployer/coremod/pkg/corebottom"
)

func BuildFrom(name string, policy corebottom.PolicyDocument) (string, error) {
	p := makePolicyJson(name, policy)

	bs, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

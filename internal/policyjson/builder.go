package policyjson

import (
	"encoding/json"

	"ziniki.org/deployer/coremod/pkg/external"
)

func BuildFrom(policy external.PolicyDocument) (string, error) {
	p := makePolicyJson(policy)

	bs, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bs), nil
}

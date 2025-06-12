package policyjson_test

import (
	"fmt"
	"testing"

	"ziniki.org/deployer/coremod/pkg/coremod"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/policyjson"
)

func TestEmptyPolicy(t *testing.T) {
	loc := &errorsink.Location{}
	doc := coremod.NewPolicyDocument(loc)
	json, err := policyjson.BuildFrom(doc)
	if err != nil {
		t.Fatalf("error generating json")
	}
	test := `{
  "Version": "2012-10-17",
  "Statement": []
}`
	if json != test {
		fmt.Printf("expected:  %s\n", test)
		fmt.Printf("generated: %s\n", json)
		t.Fatalf("incorrect json generated")
	}
}

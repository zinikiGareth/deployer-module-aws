package policyjson_test

import (
	"fmt"
	"testing"

	"ziniki.org/deployer/coremod/pkg/coremod"
	"ziniki.org/deployer/coremod/pkg/external"
	"ziniki.org/deployer/deployer/pkg/errorsink"
	"ziniki.org/deployer/modules/aws/internal/policyjson"
)

var loc *errorsink.Location = &errorsink.Location{}
var doc external.PolicyDocument = coremod.NewPolicyDocument(loc, "policy-name")

func TestEmptyPolicy(t *testing.T) {
	compare(t, doc, `{
  "Version": "2012-10-17",
  "Statement": []
}`)
}

func TestSimplePolicy(t *testing.T) {
	doc.Item("Allow")
	compare(t, doc, `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "policy-name-sid-0",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "*"
    }
  ]
}`)
}

func compare(t *testing.T, doc external.PolicyDocument, test string) {
	json, err := policyjson.BuildFrom(doc)
	if err != nil {
		t.Fatalf("error generating json")
	}
	if json != test {
		fmt.Printf("expected:  %s\n", test)
		fmt.Printf("generated: %s\n", json)
		t.Fatalf("incorrect json generated")
	}
}

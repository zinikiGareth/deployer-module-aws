package policyjson_test

import (
	"fmt"
	"strings"
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

func TestExplicitAction(t *testing.T) {
	item := doc.Item("Allow")
	item.Action("s3:*")
	compare(t, doc, `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "policy-name-sid-0",
      "Effect": "Allow",
      "Action": "s3:*",
      "Resource": "*"
    }
  ]
}`)
}

func TestExplicitResource(t *testing.T) {
	item := doc.Item("Allow")
	item.Resource("arn:s3:*")
	compare(t, doc, `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "policy-name-sid-0",
      "Effect": "Allow",
      "Action": "*",
      "Resource": "arn:s3:*"
    }
  ]
}`)
}

func TestMultipleActionsAndResources(t *testing.T) {
	item := doc.Item("Allow")
	item.Action("s3:*")
	item.Action("iot:*")
	item.Resource("arn:s3:*")
	item.Resource("arn:dynamodb:*")
	compare(t, doc, `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "policy-name-sid-0",
      "Effect": "Allow",
      "Action": [
	    "s3:*",
		"iot:*"
	  ],
      "Resource": [
	    "arn:s3:*",
		"arn:dynamodb:*"
	  ]
    }
  ]
}`)
}

func compare(t *testing.T, doc external.PolicyDocument, test string) {
	// expand tabs
	// this is not perfect, but should fix 90% of tabbing issues
	// remember that you are no worse off with the other 10%
	test = strings.ReplaceAll(test, "\t", "    ")

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

package policyjson_test

import (
	"testing"
)

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

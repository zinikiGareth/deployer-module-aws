package iam

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type policyFinder struct {
	tools *corebottom.Tools

	loc  *errorsink.Location
	name string
	coin corebottom.CoinId

	// client        *s3.Client
	// alreadyExists bool
	// cloud *BucketCloud
}

func (b *policyFinder) Loc() *errorsink.Location {
	return b.loc
}

func (b *policyFinder) ShortDescription() string {
	return "aws.IAM.Policy[" + b.name + "]"
}

func (b *policyFinder) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("aws.IAM.Policy[")
	iw.AttrsWhere(b)
	iw.TextAttr("named", b.name)
	iw.EndAttrs()
}

func (b *policyFinder) CoinId() corebottom.CoinId {
	return b.coin
}

func (b *policyFinder) DetermineInitialState(pres corebottom.ValuePresenter) {
	/*
		eq := b.tools.Recall.ObtainDriver("aws.AwsEnv")
		awsEnv, ok := eq.(*env.AwsEnv)
		if !ok {
			panic("could not cast env to AwsEnv")
		}

		b.client = awsEnv.S3Client()
		_, err := b.client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
			Bucket: aws.String(b.name),
		})
		if err != nil {
			var api smithy.APIError
			if e.As(err, &api) {
				// log.Printf("code: %s", api.ErrorCode())
				if api.ErrorCode() == "NotFound" {
					log.Printf("bucket does not exist: %s", b.name)
				} else {
					log.Fatal(err)
				}
			} else {
				log.Fatal(err)
			}
		} else {
			log.Printf("bucket exists: %s", b.name)
			b.alreadyExists = true
		}

		// TODO: I think we should capture the cloud bucket
		// (but maybe not, maybe the name is all we need)
		pres.Present(b)
	*/
}

func (b *policyFinder) String() string {
	return fmt.Sprintf("FindPolicy[%s:%s]", "" /* eb.env.Region */, b.name)
}

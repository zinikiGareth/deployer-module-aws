package iam

import (
	"context"
	"log"
	"reflect"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/coremod/pkg/corepkg"
	"ziniki.org/deployer/modules/aws/internal/env"
)

type UserFinder struct {
	Client *iam.Client
}

func (u *UserFinder) DetermineInitialState(creator *corepkg.CoreCreator, pres corebottom.ValuePresenter) {
	creator.GetEnv("aws.AwsEnv", reflect.TypeFor[*env.AwsEnv](), "IAMClient", "Client")
	log.Printf("client = %T %p\n", u.Client, u.Client)
	model := u.findUserNamed(creator.Name())
	if model == nil {
		log.Printf("user %s not found\n", creator.Name())
		pres.NotFound()
	} else {
		log.Printf("user found for %s\n", creator.Name())
		pres.Present(model)
	}
}

type userAWSModel struct {
	arn  string
	name string
}

func (cc *UserFinder) findUserNamed(name string) *userAWSModel {
	user, err := cc.Client.GetUser(context.TODO(), &iam.GetUserInput{UserName: &name})
	if err != nil {
		panic(err)
	}
	if user != nil {
		return &userAWSModel{arn: *user.User.Arn, name: name}
	} else {
		return nil
	}
}

var _ corepkg.FindStrategy = &UserFinder{}

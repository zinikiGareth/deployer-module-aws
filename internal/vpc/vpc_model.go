package vpc

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type vpcAWSModel struct {
	loc            *errorsink.Location
	vpc            *types.Vpc
	subnets        []string
	securityGroups []string
}

func (model *vpcAWSModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "id":
		return &vpcIdMethod{}
	case "subnets":
		return &subnetsMethod{}
	case "securityGroups":
		return &sgMethod{}
	}
	return nil
}

type vpcIdMethod struct {
}

func (a *vpcIdMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*vpcAWSModel)
	if !ok {
		panic(fmt.Sprintf("zoneId can only be called on a domain, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	return model.vpc.VpcId
}

type subnetsMethod struct {
}

func (a *subnetsMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*vpcAWSModel)
	if !ok {
		panic(fmt.Sprintf("zoneId can only be called on a domain, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	return model.subnets
}

type sgMethod struct {
}

func (a *sgMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	model, ok := e.(*vpcAWSModel)
	if !ok {
		panic(fmt.Sprintf("zoneId can only be called on a domain, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	return model.securityGroups
}

var _ driverbottom.HasMethods = &vpcAWSModel{}

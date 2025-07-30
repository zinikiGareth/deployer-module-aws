package iam

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
)

type RoleModel struct {
	name string

	managed []driverbottom.Expr
	inline  []corebottom.PolicyActionList
}

type RoleAWSModel struct {
	role     *types.Role
	policies []string
}

func (r *RoleAWSModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "arn":
		return &arnMethod{me: r.role}
	default:
		log.Fatalf("there is no method %s on Role", name)
		return nil
	}
}

type arnMethod struct {
	me *types.Role
}

func (a *arnMethod) Invoke(storage driverbottom.RuntimeStorage, obj driverbottom.Expr, args []driverbottom.Expr) any {
	if len(args) != 0 {
		return fmt.Errorf("arn method does not take arguments")
	}
	return *a.me.Arn
}

var _ driverbottom.Method = &arnMethod{}
var _ driverbottom.HasMethods = &RoleAWSModel{}

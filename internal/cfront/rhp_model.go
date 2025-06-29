package cfront

import (
	"fmt"

	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type rhpModel struct {
	loc    *errorsink.Location
	name   string
	header driverbottom.Expr
	value  driverbottom.Expr
	rpId   string
}

func (rm *rhpModel) Loc() *errorsink.Location {
	return rm.loc
}

func (rm *rhpModel) ShortDescription() string {
	return "RHPModel[" + rm.name + "]"
}

func (rm *rhpModel) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("RHPModel %s", rm.name)
	iw.AttrsWhere(rm)
	iw.NestedAttr("header", rm.header)
	iw.NestedAttr("value", rm.value)
	iw.EndAttrs()
}

func (rm *rhpModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "id":
		return &RHPIdMethod{}
	}
	return nil
}

type RHPIdMethod struct {
}

func (rmm *RHPIdMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	m, ok := e.(*rhpModel)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a RHP, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if m.rpId != "" {
		return m.rpId
	} else {
		return utils.DeferString(func() string {
			if m.rpId == "" {
				panic("id is still not set")
			}
			return m.rpId
		})
	}
}

var _ driverbottom.Describable = &rhpModel{}
var _ driverbottom.HasMethods = &rhpModel{}

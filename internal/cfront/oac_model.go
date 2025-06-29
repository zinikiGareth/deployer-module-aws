package cfront

import (
	"fmt"

	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type oacModel struct {
	loc  *errorsink.Location
	name string

	acType       driverbottom.Expr
	signBehavior driverbottom.Expr
	signProt     driverbottom.Expr

	oacId string
}

func (m *oacModel) Loc() *errorsink.Location {
	return m.loc
}

func (m *oacModel) ShortDescription() string {
	return fmt.Sprintf("OACModel[" + m.name + "]")
}

func (m *oacModel) DumpTo(iw driverbottom.IndentWriter) {
	iw.Intro("OACModel %s", m.name)
	iw.AttrsWhere(m)
	if m.oacId != "" {
		iw.TextAttr("oacId", m.oacId)
	}
	iw.EndAttrs()
}

func (m *oacModel) ObtainMethod(name string) driverbottom.Method {
	switch name {
	case "id":
		return &oacIdMethod{}
	}
	return nil
}

type oacIdMethod struct {
}

func (a *oacIdMethod) Invoke(s driverbottom.RuntimeStorage, on driverbottom.Expr, args []driverbottom.Expr) any {
	e := on.Eval(s)
	cfdc, ok := e.(*oacModel)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a OAC, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if cfdc.oacId != "" {
		return cfdc.oacId
	} else {
		return utils.DeferString(func() string {
			if cfdc.oacId == "" {
				panic("id is still not set")
			}
			return cfdc.oacId
		})
	}
}

var _ driverbottom.Describable = &oacModel{}
var _ driverbottom.HasMethods = &oacModel{}

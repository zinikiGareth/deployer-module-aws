package cfront

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
	"ziniki.org/deployer/driver/pkg/utils"
)

type oacModel struct {
	loc  *errorsink.Location
	name string
	coin corebottom.CoinId

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
	model, ok := e.(*oacModel)
	if !ok {
		panic(fmt.Sprintf("arn can only be called on a OAC, not a %T", e))
	}
	if len(args) != 0 {
		panic("invalid number of arguments")
	}
	if model.oacId != "" {
		return model.oacId
	} else {
		return utils.DeferString(func() string {
			curr := s.GetCoinFrom(model.coin, []int{1, 3})
			if curr == nil {
				panic("could not find find/create version of " + model.coin.VarName().Id())
			}

			oac := curr.(*oacModel)
			if oac.oacId == "" {
				panic("oac id is still not set")
			}
			return oac.oacId
		})
	}
}

var _ driverbottom.Describable = &oacModel{}
var _ driverbottom.HasMethods = &oacModel{}

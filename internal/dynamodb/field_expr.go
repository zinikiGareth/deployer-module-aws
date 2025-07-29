package dynamodb

import (
	"fmt"

	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type DynamoFieldExpr struct {
	loc     *errorsink.Location
	name    string
	ftype   string
	keytype string
}

func (f *DynamoFieldExpr) SetKeyType(tools *driverbottom.CoreTools, id driverbottom.Identifier) {
	if f.keytype != "" {
		tools.Reporter.ReportAtf(id.Loc(), "cannot set @Key multiple times on field %s", f.name)
	}
	f.keytype = id.Id()
}

func (f *DynamoFieldExpr) Loc() *errorsink.Location {
	return f.loc
}

func (f *DynamoFieldExpr) ShortDescription() string {
	return fmt.Sprintf("Field[%s %s]", f.name, f.ftype)
}

func (f *DynamoFieldExpr) DumpTo(to driverbottom.IndentWriter) {
	to.Intro("DynamoFieldExpr")
	to.AttrsWhere(f)
	to.TextAttr("name", f.name)
	to.TextAttr("type", f.ftype)
	to.EndAttrs()
}

func (f *DynamoFieldExpr) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	return driverbottom.MAY_BE_BOUND
}

func (f *DynamoFieldExpr) Eval(s driverbottom.RuntimeStorage) any {
	// I think for all intents and purposes this is evaluated already
	return f
}

func (f *DynamoFieldExpr) String() string {
	return f.ShortDescription()
}

var _ driverbottom.Expr = &DynamoFieldExpr{}

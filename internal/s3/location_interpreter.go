package s3

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

type S3Location struct {
	tools *corebottom.Tools
	driverbottom.Locatable
	Bucket driverbottom.Expr
	Key    driverbottom.Expr
}

func (s *S3Location) AddAdverb(adverb driverbottom.Adverb, args []driverbottom.Token) driverbottom.Interpreter {
	panic("unimplemented")
}

func (s *S3Location) AddProperty(name driverbottom.Identifier, expr driverbottom.Expr) {
	switch name.Id() {
	case "Bucket":
		if s.Bucket != nil {
			s.tools.Reporter.ReportAtf(name.Loc(), "cannot set %s on an S3Location multiple times", name.Id())
			return
		}
		s.Bucket = expr
	case "Key":
		if s.Key != nil {
			s.tools.Reporter.ReportAtf(name.Loc(), "cannot set %s on an S3Location multiple times", name.Id())
			return
		}
		s.Key = expr
	default:
		s.tools.Reporter.ReportAtf(name.Loc(), "there is no property %s on an S3Location", name.Id())
	}
}

func (s *S3Location) Completed() {
	if s.Bucket == nil {
		s.tools.Reporter.ReportAtf(s.Loc(), "must set Bucket on an S3Location")
	}
	if s.Key == nil {
		s.tools.Reporter.ReportAtf(s.Loc(), "must set Key on an S3Location")
	}
}

func (s *S3Location) ShortDescription() string {
	panic("unimplemented")
}

func (s *S3Location) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

func (s *S3Location) String() string {
	return fmt.Sprintf("S3[%s:%s]", s.Bucket.String(), s.Key.String())
}

func (*S3Location) Eval(s driverbottom.RuntimeStorage) any {
	panic("unimplemented")
}

func (s *S3Location) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	ret := driverbottom.MAY_BE_BOUND
	ret = ret.Merge(s.Bucket.Resolve(r))
	ret = ret.Merge(s.Key.Resolve(r))
	return ret
}

func CreateLocationInterpreter(tools *driverbottom.CoreTools, scope driverbottom.Scope, parent driverbottom.PropertyParent, prop driverbottom.Identifier, tokens []driverbottom.Token) driverbottom.Interpreter {
	s3loc := &S3Location{Locatable: prop}
	parent.AddProperty(prop, s3loc)
	return drivertop.NewPropertiesInnerScope(tools, s3loc)
}

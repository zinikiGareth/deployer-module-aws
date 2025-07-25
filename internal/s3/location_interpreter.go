package s3

import (
	"fmt"
	"log"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
	"ziniki.org/deployer/driver/pkg/errorsink"
)

type S3Location struct {
	tools *corebottom.Tools
	driverbottom.Locatable
	Bucket driverbottom.Expr
	Key    driverbottom.Expr
}

// AddAdverb implements driverbottom.PropertyParent.
func (s *S3Location) AddAdverb(adverb driverbottom.Adverb, args []driverbottom.Token) driverbottom.Interpreter {
	panic("unimplemented")
}

// AddProperty implements driverbottom.PropertyParent.
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

// Completed implements driverbottom.PropertyParent.
func (s *S3Location) Completed() {
	if s.Bucket == nil {
		s.tools.Reporter.ReportAtf(s.Loc(), "must set Bucket on an S3Location")
	}
	if s.Key == nil {
		s.tools.Reporter.ReportAtf(s.Loc(), "must set Key on an S3Location")
	}
}

// DumpTo implements driverbottom.Expr.
func (s *S3Location) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

// Eval implements driverbottom.Expr.
func (*S3Location) Eval(s driverbottom.RuntimeStorage) any {
	panic("unimplemented")
}

// Loc implements driverbottom.Expr.
func (s *S3Location) Loc() *errorsink.Location {
	panic("unimplemented")
}

// Resolve implements driverbottom.Expr.
func (s *S3Location) Resolve(r driverbottom.Resolver) {
	s.Bucket.Resolve(r)
	s.Key.Resolve(r)
}

// ShortDescription implements driverbottom.Expr.
func (s *S3Location) ShortDescription() string {
	panic("unimplemented")
}

// String implements driverbottom.Expr.
func (s *S3Location) String() string {
	return fmt.Sprintf("S3[%s:%s]", s.Bucket.String(), s.Key.String())
}

func CreateLocationInterpreter(tools *driverbottom.CoreTools, parent driverbottom.PropertyParent, prop driverbottom.Identifier) driverbottom.Interpreter {
	log.Printf("have %T %v %s\n", parent, parent, prop.Id())
	s3loc := &S3Location{Locatable: prop}
	parent.AddProperty(prop, s3loc)
	return drivertop.NewPropertiesInnerScope(tools, s3loc)
}

// var _ driverbottom.Interpreter = &S3LocationInterpreter{}

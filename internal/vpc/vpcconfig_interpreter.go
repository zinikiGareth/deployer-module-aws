package vpc

import (
	"fmt"

	"ziniki.org/deployer/coremod/pkg/corebottom"
	"ziniki.org/deployer/driver/pkg/driverbottom"
	"ziniki.org/deployer/driver/pkg/drivertop"
)

type VPCConfig struct {
	tools *corebottom.Tools
	driverbottom.Locatable
	DualStack      driverbottom.Expr
	Subnets        driverbottom.Expr
	SecurityGroups driverbottom.Expr
}

func (s *VPCConfig) AddAdverb(adverb driverbottom.Adverb, args []driverbottom.Token) driverbottom.Interpreter {
	panic("unimplemented")
}

func (s *VPCConfig) AddProperty(name driverbottom.Identifier, expr driverbottom.Expr) {
	switch name.Id() {
	case "DualStack":
		if s.DualStack != nil {
			s.tools.Reporter.ReportAtf(name.Loc(), "cannot set %s on an VPCConfig multiple times", name.Id())
			return
		}
		s.DualStack = expr
	case "Subnets":
		if s.Subnets != nil {
			s.tools.Reporter.ReportAtf(name.Loc(), "cannot set %s on an VPCConfig multiple times", name.Id())
			return
		}
		s.Subnets = expr
	case "SecurityGroups":
		if s.SecurityGroups != nil {
			s.tools.Reporter.ReportAtf(name.Loc(), "cannot set %s on an VPCConfig multiple times", name.Id())
			return
		}
		s.SecurityGroups = expr
	default:
		s.tools.Reporter.ReportAtf(name.Loc(), "there is no property %s on an VPCConfig", name.Id())
	}
}

func (s *VPCConfig) Completed() {
	if s.Subnets == nil {
		s.tools.Reporter.ReportAtf(s.Loc(), "must set Subnets on an VPCConfig")
	}
	if s.SecurityGroups == nil {
		s.tools.Reporter.ReportAtf(s.Loc(), "must set SecurityGroups on an VPCConfig")
	}
}

func (s *VPCConfig) ShortDescription() string {
	panic("unimplemented")
}

func (s *VPCConfig) DumpTo(to driverbottom.IndentWriter) {
	panic("unimplemented")
}

func (s *VPCConfig) String() string {
	return fmt.Sprintf("S3[%s:%s]", s.Subnets.String(), s.SecurityGroups.String())
}

func (s *VPCConfig) Resolve(r driverbottom.Resolver) driverbottom.BindingRequirement {
	ret := driverbottom.MAY_BE_BOUND
	if s.DualStack != nil {
		ret = ret.Merge(s.DualStack.Resolve(r))
	}
	ret = ret.Merge(s.Subnets.Resolve(r))
	ret = ret.Merge(s.SecurityGroups.Resolve(r))
	return ret
}

func (v *VPCConfig) Eval(s driverbottom.RuntimeStorage) any {
	ret := make(map[string][]string)
	if v.DualStack != nil {
		b := s.Eval(v.DualStack)
		if b != nil {
			b1, ok := b.(float64)
			if ok && b1 == 1 {
				ret["DualStack"] = []string{}
			}
		}
	}
	ret["Subnets"] = v.Subnets.Eval(s).([]string)
	ret["SecurityGroups"] = v.SecurityGroups.Eval(s).([]string)
	return ret
}

func CreateConfigInterpreter(tools *driverbottom.CoreTools, scope driverbottom.Scope, parent driverbottom.PropertyParent, prop driverbottom.Identifier, tokens []driverbottom.Token) driverbottom.Interpreter {
	s3loc := &VPCConfig{tools: tools.RetrieveOther("coremod").(*corebottom.Tools), Locatable: prop}
	parent.AddProperty(prop, s3loc)
	return drivertop.NewPropertiesInnerScope(tools, s3loc)
}

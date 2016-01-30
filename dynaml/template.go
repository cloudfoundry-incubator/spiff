package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type TemplateExpr struct {
}

func (e TemplateExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info := DefaultInfo()
	info.Issue = yaml.NewIssue("&template only usable to declare templates")
	return nil, info, false
}

func (e TemplateExpr) String() string {
	return fmt.Sprintf("&template")
}

type SubstitutionExpr struct {
	Template Expression
	Node     yaml.Node
}

func (e SubstitutionExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	if e.Node == nil {
		debug.Debug("evaluating expression to determine template\n")
		n, info, ok := e.Template.Evaluate(binding)
		if !ok || isExpression((n)) {
			return node(e), info, ok
		}
		val, ok := n.Value().(TemplateValue)
		if !ok {
			info.Issue = yaml.NewIssue("template value required")
			return nil, info, false
		} else {
			e.Node = val.Prepared
		}
	}
	debug.Debug("resolving template\n")
	result, state := binding.Flow(e.Node, false)
	info := DefaultInfo()
	if state != nil {
		debug.Debug("resolving template failed: " + state.Error())
		info.Issue = state.Issue("template resolution failed")
		return node(e), info, true
	}
	debug.Debug("resolving template succeeded")
	return result, info, true
}

func (e SubstitutionExpr) String() string {
	return fmt.Sprintf("*(%s)", e.Template)
}

type TemplateValue struct {
	Prepared yaml.Node
	Orig     yaml.Node
}

func (e TemplateValue) MarshalYAML() (tag string, value interface{}) {
	return e.Orig.MarshalYAML()
}

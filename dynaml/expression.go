package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type Binding interface {
	FindFromRoot([]string) (yaml.Node, bool)
	FindReference([]string) (yaml.Node, bool)
	FindInStubs([]string) (yaml.Node, bool)
}

type EvaluationInfo struct {
	RedirectPath []string
	Replace      bool
	Merged       bool
	Preferred    bool
}

func DefaultInfo() EvaluationInfo {
	return EvaluationInfo{nil,false,false,false}
}

type Expression interface {
	Evaluate(Binding) (yaml.Node, EvaluationInfo, bool)
}

func (i EvaluationInfo) Join(o EvaluationInfo) EvaluationInfo {
	if o.RedirectPath !=nil {
		i.RedirectPath = o.RedirectPath
	}
	i.Replace = o.Replace // replace only by directly using the merge node
	i.Preferred = i.Preferred || o.Preferred
	return i
}


func ResolveIntegerExpressionOrPushEvaluation(e *Expression, resolved *bool, info *EvaluationInfo, binding Binding) (int64, EvaluationInfo, bool) {
	node, infoe, ok := (*e).Evaluate(binding)
	if info!=nil {
		infoe = (*info).Join(infoe)
	}
	if !ok {
		return 0, infoe, false
	}

	switch node.Value().(type) {
		case Expression:
			*e = node.Value().(Expression)
			*resolved = false
			return 0, infoe, true
		case int64:
			return node.Value().(int64), infoe, true
		default:
			return 0, infoe, false
	}
}
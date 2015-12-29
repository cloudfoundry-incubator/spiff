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

func ResolveExpressionOrPushEvaluation(e *Expression, resolved *bool, info *EvaluationInfo, binding Binding) (interface{}, EvaluationInfo, bool) {
	node, infoe, ok := (*e).Evaluate(binding)
	if info!=nil {
		infoe = (*info).Join(infoe)
	}
	if !ok {
		return nil, infoe, false
	}

	switch node.Value().(type) {
		case Expression:
			*e = node.Value().(Expression)
			*resolved = false
			return nil, infoe, true
		default:
			return node.Value(), infoe, true
	}
}

func ResolveIntegerExpressionOrPushEvaluation(e *Expression, resolved *bool, info *EvaluationInfo, binding Binding) (int64, EvaluationInfo, bool) {
	value, infoe, ok := ResolveExpressionOrPushEvaluation(e,resolved,info,binding)

    if value == nil {
		return 0, infoe, ok
	}
	
    i, ok := value.(int64)
	if ok {
		return i, infoe, true
	} else {
		return 0, infoe, false
	}
}

func ResolveExpressionListOrPushEvaluation(list *[]Expression, resolved *bool, info *EvaluationInfo, binding Binding) ([]interface{}, EvaluationInfo, bool) {
	values := make([]interface{}, len(*list))
	pushed := make([]Expression, len(*list))
	infoe  := EvaluationInfo{}
	ok     := true
	
	copy(pushed,*list)
	
	for i, _ := range pushed {
		values[i], infoe, ok = ResolveExpressionOrPushEvaluation(&pushed[i],resolved,info,binding)
		info=&infoe
		if !ok {
		  return nil, infoe, false
		}
	}
	*list=pushed
	return values, infoe, true
	
}
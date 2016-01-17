package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func (e CallExpr) defined(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	pushed := make([]Expression, len(e.Arguments))
	ok := true
	resolved := true

	copy(pushed, e.Arguments)
	for i, _ := range pushed {
		_, _, ok = ResolveExpressionOrPushEvaluation(&pushed[i], &resolved, nil, binding)
		if resolved && !ok {
			return node(false), DefaultInfo(), true
		}
	}
	if !resolved {
		return node(e), DefaultInfo(), true
	}
	return node(true), DefaultInfo(), ok
}

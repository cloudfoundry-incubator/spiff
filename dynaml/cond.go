package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type CondExpr struct {
	C Expression
	T Expression
	F Expression
}

func (e CondExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	resolved := true
	info := DefaultInfo()
	var result yaml.Node

	a, info, ok := ResolveExpressionOrPushEvaluation(&e.C, &resolved, &info, binding)
	if !ok {
		return nil, info, false
	}
	if !resolved {
		return node(e), info, true
	}
	if toBool(a) {
		result, info, ok = e.T.Evaluate(binding)
	} else {
		result, info, ok = e.F.Evaluate(binding)
	}
	return result, info, ok
}

func (e CondExpr) String() string {
	return fmt.Sprintf("%s ? %s : %s", e.C, e.T, e.F)
}

func toBool(v interface{}) bool {
	if v == nil {
		return false
	}

	switch eff := v.(type) {
	case bool:
		return eff
	case string:
		return len(eff) > 0
	case int64:
		return eff != 0
	case []yaml.Node:
		return len(eff) != 0
	case map[string]yaml.Node:
		return len(eff) != 0
	default:
		return true
	}
}

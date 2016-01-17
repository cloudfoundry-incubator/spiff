package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type RangeExpr struct {
	Start Expression
	End   Expression
}

func (e RangeExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	resolved := true

	start, info, ok := ResolveIntegerExpressionOrPushEvaluation(&e.Start, &resolved, nil, binding)
	if !ok {
		return nil, info, false
	}
	end, info, ok := ResolveIntegerExpressionOrPushEvaluation(&e.End, &resolved, &info, binding)
	if !ok {
		return nil, info, false
	}
	if !resolved {
		return node(e), info, true
	}

	nodes := []yaml.Node{}
	delta := int64(1)
	if start > end {
		delta = -1
	}
	for i := start; i*delta <= end*delta; i += delta {
		nodes = append(nodes, node(i))
	}

	return node(nodes), info, true
}

func (e RangeExpr) String() string {
	return fmt.Sprintf("[%s..%s]", e.Start, e.End)
}

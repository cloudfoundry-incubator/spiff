package dynaml

import (
	"fmt"
)

type PreferExpr struct {
	expression Expression
}

func (e PreferExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {

	val, info, ok := e.expression.Evaluate(binding)
	info.Preferred = true
	return val, info, ok
}

func (e PreferExpr) String() string {
	return fmt.Sprintf("prefer %s", e.expression)
}

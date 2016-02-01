package dynaml

import (
	"fmt"
)

type StringExpr struct {
	Value string
}

func (e StringExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
	return e.Value, DefaultInfo(), true
}

func (e StringExpr) String() string {
	return fmt.Sprintf("%q", e.Value)
}

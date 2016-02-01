package dynaml

import (
	"strconv"
)

type IntegerExpr struct {
	Value int64
}

func (e IntegerExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
	return e.Value, DefaultInfo(), true
}

func (e IntegerExpr) String() string {
	return strconv.FormatInt(e.Value, 10)
}

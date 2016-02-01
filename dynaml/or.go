package dynaml

import (
	"fmt"
	"reflect"
)

type OrExpr struct {
	A Expression
	B Expression
}

func (e OrExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
	a, infoa, ok := e.A.Evaluate(binding)
	if ok {
		if reflect.DeepEqual(a, e.A) {
			return nil, infoa, false
		}
		return a, infoa, true
	}

	b, infob, ok := e.B.Evaluate(binding)
	return b, infoa.Join(infob), ok
}

func (e OrExpr) String() string {
	return fmt.Sprintf("%s || %s", e.A, e.B)
}

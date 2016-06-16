package dynaml

import (
	"fmt"
)

type MultiplicationExpr struct {
	A Expression
	B Expression
}

func (e MultiplicationExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
	resolved := true

	aint, info, ok := ResolveIntegerExpressionOrPushEvaluation(&e.A, &resolved, nil, binding)
	if !ok {
		return nil, info, false
	}

	bint, info, ok := ResolveIntegerExpressionOrPushEvaluation(&e.B, &resolved, &info, binding)
	if !ok {
		return nil, info, false
	}

	if !resolved {
		return e, info, true
	}
	return aint * bint, info, true
}

func (e MultiplicationExpr) String() string {
	return fmt.Sprintf("%s * %s", e.A, e.B)
}

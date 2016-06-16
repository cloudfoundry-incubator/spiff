package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type DivisionExpr struct {
	A Expression
	B Expression
}

func (e DivisionExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
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

	if bint == 0 {
		info.Issue = yaml.NewIssue("division by zero")
		return nil, info, false
	}
	return aint / bint, info, true
}

func (e DivisionExpr) String() string {
	return fmt.Sprintf("%s / %s", e.A, e.B)
}

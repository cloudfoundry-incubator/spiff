package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/debug"
)

const (
	OpAnd = "-and"
)

type LogAndExpr struct {
	A Expression
	B Expression
}

func (e LogAndExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
	a, b, info, resolved, ok := resolveLOperands(e.A, e.B, binding)
	if !ok {
		return nil, info, false
	}
	if !resolved {
		return e, info, true
	}
	debug.Debug("AND: %#v, %#v\n", a, b)
	inta, ok := a.(int64)
	if ok {
		return inta & b.(int64), info, true
	}
	return toBool(a) && toBool(b), info, true
}

func (e LogAndExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.A, OpAnd, e.B)
}

package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/debug"
)

const (
	OpOr = "-or"
)

type LogOrExpr struct {
	A Expression
	B Expression
}

func (e LogOrExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
	a, b, info, resolved, ok := resolveLOperands(e.A, e.B, binding)
	if !ok {
		debug.Debug("OR: failed %#v, %#v\n", e.A, e.B)
		return nil, info, false
	}
	if !resolved {
		return e, info, true
	}
	debug.Debug("OR: %#v, %#v\n", a, b)
	inta, ok := a.(int64)
	if ok {
		return inta | b.(int64), info, true
	}
	return (toBool(a) || toBool(b)), info, true
}

func (e LogOrExpr) String() string {
	return fmt.Sprintf("%s %s %s", e.A, OpOr, e.B)
}

func resolveLOperands(a, b Expression, binding Binding) (eff_a, eff_b interface{}, info EvaluationInfo, resolved bool, ok bool) {
	var va, vb interface{}
	var infoa, infob EvaluationInfo

	va, infoa, ok = a.Evaluate(binding)
	if ok {
		if isExpression(va) {
			return nil, nil, infoa, false, true
		}

		vb, infob, ok = b.Evaluate(binding)
		info = infoa.Join(infob)
		if !ok {
			return nil, nil, info, false, false
		}

		if isExpression(vb) {
			return nil, nil, info, false, true
		}

		resolved = true
		eff_a, ok = va.(int64)
		if ok {
			eff_b, ok = vb.(int64)
			if ok {
				return
			}
		}
		return toBool(va), toBool(vb), info, true, true
	}

	return nil, nil, info, false, false
}

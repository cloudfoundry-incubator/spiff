package dynaml

import (
	"fmt"

	"github.com/shutej/spiff/yaml"
)

type AdditionExpr struct {
	A Expression
	B Expression
}

func (e AdditionExpr) RequiresPhases(binding Binding) StringSet {
	return e.A.RequiresPhases(binding).Union(e.B.RequiresPhases(binding))
}

func (e AdditionExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	a, ok := e.A.Evaluate(binding)
	if !ok {
		return nil, false
	}

	b, ok := e.B.Evaluate(binding)
	if !ok {
		return nil, false
	}

	aint, ok := a.Value().(int64)
	if !ok {
		return nil, false
	}

	bint, ok := b.Value().(int64)
	if !ok {
		return nil, false
	}

	return node(aint + bint), true
}

func (e AdditionExpr) String() string {
	return fmt.Sprintf("%s + %s", e.A, e.B)
}

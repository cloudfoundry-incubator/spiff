package dynaml

import (
	"fmt"

	"github.com/shutej/spiff/yaml"
)

type SubtractionExpr struct {
	A Expression
	B Expression
}

func (e SubtractionExpr) RequiresPhases() StringSet {
	return StringSet(nil)
}

func (e SubtractionExpr) Evaluate(binding Binding) (yaml.Node, bool) {
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

	return node(aint - bint), true
}

func (e SubtractionExpr) String() string {
	return fmt.Sprintf("%s - %s", e.A, e.B)
}

package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type SubtractionExpr struct {
	A Expression
	B Expression
}

func (e SubtractionExpr) Evaluate(context Context) (yaml.Node, bool) {
	a, ok := e.A.Evaluate(context)
	if !ok {
		return nil, false
	}

	b, ok := e.B.Evaluate(context)
	if !ok {
		return nil, false
	}

	aint, ok := a.(int)
	if !ok {
		return nil, false
	}

	bint, ok := b.(int)
	if !ok {
		return nil, false
	}

	return aint - bint, true
}

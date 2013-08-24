package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type SubtractionExpr struct {
	A Expression
	B Expression
}

func (e SubtractionExpr) Evaluate(context Context) yaml.Node {
	a := e.A.Evaluate(context)
	b := e.B.Evaluate(context)

	aint, ok := a.(int)
	if !ok {
		return nil
	}

	bint, ok := b.(int)
	if !ok {
		return nil
	}

	return aint - bint
}

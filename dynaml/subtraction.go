package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type SubtractionExpr struct {
	A Expression
	B Expression
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

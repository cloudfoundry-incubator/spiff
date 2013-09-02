package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type ConcatenationExpr struct {
	A Expression
	B Expression
}

func (e ConcatenationExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	a, ok := e.A.Evaluate(binding)
	if !ok {
		return nil, false
	}

	b, ok := e.B.Evaluate(binding)
	if !ok {
		return nil, false
	}

	astring, ok := a.(string)
	if !ok {
		return nil, false
	}

	bstring, ok := b.(string)
	if !ok {
		return nil, false
	}

	return astring + bstring, true
}

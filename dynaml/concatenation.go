package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type ConcatenationExpr struct {
	A Expression
	B Expression
}

func (e ConcatenationExpr) Evaluate(context Context) (yaml.Node, bool) {
	a, ok := e.A.Evaluate(context)
	if !ok {
		return nil, false
	}

	b, ok := e.B.Evaluate(context)
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

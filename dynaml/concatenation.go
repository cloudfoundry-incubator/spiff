package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type ConcatenationExpr struct {
	A Expression
	B Expression
}

func (e ConcatenationExpr) Evaluate(context Context) yaml.Node {
	a := e.A.Evaluate(context)
	b := e.B.Evaluate(context)

	astring, ok := a.(string)
	if !ok {
		return nil
	}

	bstring, ok := b.(string)
	if !ok {
		return nil
	}

	return astring + bstring
}

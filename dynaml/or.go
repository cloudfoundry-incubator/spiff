package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type OrExpr struct {
	A Expression
	B Expression
}

func (e OrExpr) Evaluate(context Context) (yaml.Node, bool) {
	a, ok := e.A.Evaluate(context)
	if ok {
		return a, true
	}

	return e.B.Evaluate(context)
}

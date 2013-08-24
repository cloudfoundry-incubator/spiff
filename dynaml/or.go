package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type OrExpr struct {
	A Expression
	B Expression
}

func (e OrExpr) Evaluate(context Context) yaml.Node {
	a := e.A.Evaluate(context)
	if a != nil {
		return a
	}

	return e.B.Evaluate(context)
}

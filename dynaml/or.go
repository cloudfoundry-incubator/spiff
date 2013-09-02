package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type OrExpr struct {
	A Expression
	B Expression
}

func (e OrExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	a, ok := e.A.Evaluate(binding)
	if ok {
		return a, true
	}

	return e.B.Evaluate(binding)
}

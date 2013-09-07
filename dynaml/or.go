package dynaml

import (
	"reflect"

	"github.com/vito/spiff/yaml"
)

type OrExpr struct {
	A Expression
	B Expression
}

func (e OrExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	a, ok := e.A.Evaluate(binding)
	if ok {
		if reflect.DeepEqual(a, e.A) {
			return nil, false
		}

		return a, true
	}

	return e.B.Evaluate(binding)
}

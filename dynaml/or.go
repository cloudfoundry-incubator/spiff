package dynaml

import (
	"fmt"
	"reflect"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type OrExpr struct {
	A Expression
	B Expression
}

func (e OrExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	a, ok := e.A.Evaluate(binding)
	if ok {
		if reflect.DeepEqual(a.Value(), e.A) {
			return nil, false
		}

		return a, true
	}

	return e.B.Evaluate(binding)
}

func (e OrExpr) String() string {
	return fmt.Sprintf("%s || %s", e.A, e.B)
}

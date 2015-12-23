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

func (e OrExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	a, infoa, ok := e.A.Evaluate(binding)
	if ok {
		if reflect.DeepEqual(a.Value(), e.A) {
			return nil, infoa, false
		}

		return a, infoa, true
	}

	b, infob, ok := e.B.Evaluate(binding)
	return b, infoa.Join(infob), ok
}

func (e OrExpr) String() string {
	return fmt.Sprintf("%s || %s", e.A, e.B)
}

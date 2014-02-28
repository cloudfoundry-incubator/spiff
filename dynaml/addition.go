package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type AdditionExpr struct {
	A Expression
	B Expression
}

func (e AdditionExpr) Evaluate(binding Binding) (yaml.Node, bool) {
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

	return aint + bint, true
}

func (e AdditionExpr) String() string {
	return fmt.Sprintf("%s + %s", e.A, e.B)
}

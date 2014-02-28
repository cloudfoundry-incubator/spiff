package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
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

	astring, aok := a.Value().(string)
	bstring, bok := b.Value().(string)
	if aok && bok {
		return node(astring + bstring), true
	}

	alist, aok := a.Value().([]yaml.Node)
	blist, bok := b.Value().([]yaml.Node)
	if aok && bok {
		return node(append(alist, blist...)), true
	}

	return nil, false
}

func (e ConcatenationExpr) String() string {
	return fmt.Sprintf("%s %s", e.A, e.B)
}

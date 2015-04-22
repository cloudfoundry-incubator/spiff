package dynaml

import (
	"fmt"
	"strconv"

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

	val, ok := concatenateStringAndInt(a, b)
	if ok {
		return node(val), true
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

func concatenateStringAndInt(a yaml.Node, b yaml.Node) (string, bool) {
	aString, aOk := a.Value().(string)
	if aOk {
		bString, bOk := b.Value().(string)
		if bOk {
			return aString + bString, true
		} else {
			bInt, bOk := b.Value().(int64)
			if bOk {
				return aString + strconv.FormatInt(bInt, 10), true
			}
		}
	}

	return "", false
}

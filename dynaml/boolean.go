package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type BooleanExpr struct {
	Value bool
}

func (e BooleanExpr) Evaluate(Binding) (yaml.Node, bool) {
	return e.Value, true
}

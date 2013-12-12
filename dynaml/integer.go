package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type IntegerExpr struct {
	Value int
}

func (e IntegerExpr) Evaluate(Binding) (yaml.Node, bool) {
	return e.Value, true
}

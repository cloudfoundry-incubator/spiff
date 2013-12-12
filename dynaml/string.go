package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type StringExpr struct {
	Value string
}

func (e StringExpr) Evaluate(Binding) (yaml.Node, bool) {
	return e.Value, true
}

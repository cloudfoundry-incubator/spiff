package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type StringExpr struct {
	Value string
}

func (e StringExpr) Evaluate(Binding) (yaml.Node, bool) {
	return e.Value, true
}

func (e StringExpr) String() string {
	return fmt.Sprintf("%q", e.Value)
}

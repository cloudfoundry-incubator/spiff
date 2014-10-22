package dynaml

import (
	"fmt"

	"github.com/shutej/spiff/yaml"
)

type StringExpr struct {
	Value string
}

func (e StringExpr) RequiresPhases(binding Binding) StringSet {
	return StringSet(nil)
}

func (e StringExpr) Evaluate(Binding) (yaml.Node, bool) {
	return node(e.Value), true
}

func (e StringExpr) String() string {
	return fmt.Sprintf("%q", e.Value)
}

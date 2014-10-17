package dynaml

import (
	"github.com/shutej/spiff/yaml"
)

type NilExpr struct{}

func (e NilExpr) RequiresPhases(_ Binding) StringSet {
	return StringSet(nil)
}

func (e NilExpr) Evaluate(Binding) (yaml.Node, bool) {
	return node(nil), true
}

func (e NilExpr) String() string {
	return "nil"
}

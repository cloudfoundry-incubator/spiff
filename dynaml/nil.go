package dynaml

import (
	"github.com/shutej/spiff/yaml"
)

type NilExpr struct{}

func (e NilExpr) Evaluate(Binding) (yaml.Node, bool) {
	return node(nil), true
}

func (e NilExpr) String() string {
	return "nil"
}

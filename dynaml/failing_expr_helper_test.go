package dynaml

import (
	"github.com/shutej/spiff/yaml"
)

type FailingExpr struct{}

func (FailingExpr) Evaluate(Binding) (yaml.Node, bool) {
	return nil, false
}

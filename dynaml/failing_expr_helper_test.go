package dynaml

import (
	"github.com/shutej/spiff/yaml"
)

type FailingExpr struct{}

func (e FailingExpr) RequiresPhases() StringSet {
	return StringSet(nil)
}

func (FailingExpr) Evaluate(Binding) (yaml.Node, bool) {
	return nil, false
}

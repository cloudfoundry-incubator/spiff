package dynaml

import (
	"strconv"

	"github.com/shutej/spiff/yaml"
)

type IntegerExpr struct {
	Value int64
}

func (e IntegerExpr) RequiresPhases() StringSet {
	return StringSet(nil)
}

func (e IntegerExpr) Evaluate(Binding) (yaml.Node, bool) {
	return node(e.Value), true
}

func (e IntegerExpr) String() string {
	return strconv.FormatInt(e.Value, 10)
}

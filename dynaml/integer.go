package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type IntegerExpr struct {
	Value int
}

func (e IntegerExpr) Evaluate(Context) yaml.Node {
	return e.Value
}

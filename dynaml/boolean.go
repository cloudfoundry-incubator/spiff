package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type BooleanExpr struct {
	Value bool
}

func (e BooleanExpr) Evaluate(Context) yaml.Node {
	return e.Value
}

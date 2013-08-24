package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type CallExpr struct {
	Name      string
	Arguments []Expression
}

func (e CallExpr) Evaluate(Context) yaml.Node {
	return nil
}

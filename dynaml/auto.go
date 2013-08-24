package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type AutoExpr struct {
	Path []string
}

func (e AutoExpr) Evaluate(Context) yaml.Node {
	return nil
}

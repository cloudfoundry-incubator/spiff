package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type MergeExpr struct {
	Path []string
}

func (e MergeExpr) Evaluate(context Context) yaml.Node {
	return context.FindInStubs(e.Path)
}

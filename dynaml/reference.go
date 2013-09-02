package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type ReferenceExpr struct {
	Path []string
}

func (e ReferenceExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	reference, ok := binding.FindReference(e.Path)
	if !ok {
		return nil, false
	}

	switch reference.(type) {
	case Expression:
		return nil, false
	default:
		return reference, true
	}
}

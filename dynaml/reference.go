package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type ReferenceExpr struct {
	Path []string
}

func (e ReferenceExpr) Evaluate(context Context) yaml.Node {
	reference := context.FindReference(e.Path)

	switch reference.(type) {
	case Expression:
		return nil
	default:
		return reference
	}
}

package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type ListExpr struct {
	Contents []Expression
}

func (e ListExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	nodes := []yaml.Node{}

	for _, c := range e.Contents {
		result, ok := c.Evaluate(binding)
		if !ok {
			return nil, false
		}

		nodes = append(nodes, result)
	}

	return nodes, true
}

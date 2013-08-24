package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type ListExpr struct {
	Contents []Expression
}

func (e ListExpr) Evaluate(context Context) yaml.Node {
	nodes := []yaml.Node{}

	for _, c := range e.Contents {
		nodes = append(nodes, c.Evaluate(context))
	}

	return nodes
}

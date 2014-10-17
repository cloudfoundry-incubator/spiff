package dynaml

import (
	"fmt"
	"strings"

	"github.com/shutej/spiff/yaml"
)

type ListExpr struct {
	Contents []Expression
}

func (e ListExpr) RequiresPhases(binding Binding) StringSet {
	retval := StringSet{}
	for _, c := range e.Contents {
		retval.Update(c.RequiresPhases(binding))
	}
	return retval
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

	return node(nodes), true
}

func (e ListExpr) String() string {
	vals := make([]string, len(e.Contents))
	for i, e := range e.Contents {
		vals[i] = fmt.Sprintf("%s", e)
	}

	return fmt.Sprintf("[%s]", strings.Join(vals, ", "))
}

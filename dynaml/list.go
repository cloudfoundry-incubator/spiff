package dynaml

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type ListExpr struct {
	Contents []Expression
}

func (e ListExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	nodes := []yaml.Node{}
    info := EvaluationInfo{nil}
	
	for _, c := range e.Contents {
		result, _, ok := c.Evaluate(binding)
		if !ok {
			return nil, info, false
		}

		nodes = append(nodes, result)
	}

	return node(nodes), info, true
}

func (e ListExpr) String() string {
	vals := make([]string, len(e.Contents))
	for i, e := range e.Contents {
		vals[i] = fmt.Sprintf("%s", e)
	}

	return fmt.Sprintf("[%s]", strings.Join(vals, ", "))
}

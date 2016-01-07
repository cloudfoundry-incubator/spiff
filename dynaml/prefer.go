package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type PreferExpr struct {
	expression Expression
}

func (e PreferExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {

	node, info, ok := e.expression.Evaluate(binding)
	info.Preferred = true
	return node, info, ok
}

func (e PreferExpr) String() string {
	return fmt.Sprintf("prefer %s", e.expression)
}

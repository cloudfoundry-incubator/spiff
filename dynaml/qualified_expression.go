package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type QualifiedExpr struct {
	Expression Expression
	Reference  ReferenceExpr
}

func (e QualifiedExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {

	root, info, ok := e.Expression.Evaluate(binding)
	if !ok {
		return nil, info, false
	}
	if !isResolved(root) {
		return node(e), info, true
	}

	debug.Debug("qualified reference: %v\n", e.Reference.Path)
	return e.Reference.find(func(end int, path []string) (yaml.Node, bool) {
		return yaml.Find(root, e.Reference.Path[0:end+1]...)
	})
}

func (e QualifiedExpr) String() string {
	return fmt.Sprintf("(%s).%s", e.Expression, e.Reference)
}

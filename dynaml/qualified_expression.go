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

func (e QualifiedExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {

	root, info, ok := e.Expression.Evaluate(binding)
	if !ok {
		return nil, info, false
	}
	if !isResolvedValue(root) {
		return e, info, true
	}

	debug.Debug("qualified reference: %v\n", e.Reference.Path)
	return e.Reference.find(func(end int, path []string) (yaml.Node, bool) {
		return yaml.Find(node(root, nil), e.Reference.Path[0:end+1]...)
	}, binding)
}

func (e QualifiedExpr) String() string {
	return fmt.Sprintf("(%s).%s", e.Expression, e.Reference)
}

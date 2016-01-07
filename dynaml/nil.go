package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type NilExpr struct{}

func (e NilExpr) Evaluate(Binding) (yaml.Node, EvaluationInfo, bool) {
	return node(nil), DefaultInfo(), true
}

func (e NilExpr) String() string {
	return "nil"
}

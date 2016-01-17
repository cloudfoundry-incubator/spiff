package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type NotExpr struct {
	Expr Expression
}

func (e NotExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	resolved := true
	v, info, ok := ResolveExpressionOrPushEvaluation(&e.Expr, &resolved, nil, binding)
	if !ok {
		return nil, info, false
	}
	if !resolved {
		return node(e), info, true
	}

	debug.Debug("NOT: %#v\n", v)
	return node(!toBool(v)), info, true
}

func (e NotExpr) String() string {
	return fmt.Sprintf("!(%s)", e.Expr)
}

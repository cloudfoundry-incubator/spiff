package dynaml

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type CallExpr struct {
	Name      string
	Arguments []Expression
}

func (e CallExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	resolved := true

	values, info, ok := ResolveExpressionListOrPushEvaluation(&e.Arguments, &resolved, nil, binding)

	if !ok {
		return nil, info, false
	}
	if !resolved {
		return node(e), info, true
	}

	var result yaml.Node
	var sub EvaluationInfo

	switch e.Name {
	case "static_ips":
		result, sub, ok = func_static_ips(e.Arguments, binding)

	case "join":
		result, sub, ok = func_join(values, binding)

	case "exec":
		result, sub, ok = func_exec(values, binding)

	case "min_ip":
		result, sub, ok = func_minIP(values, binding)

	case "max_ip":
		result, sub, ok = func_maxIP(values, binding)

	default:
		return nil, info, false
	}

	return result, sub.Join(info), ok
}

func (e CallExpr) String() string {
	args := make([]string, len(e.Arguments))
	for i, e := range e.Arguments {
		args[i] = fmt.Sprintf("%s", e)
	}

	return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))
}

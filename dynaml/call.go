package dynaml

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type CallExpr struct {
	Function  Expression
	Arguments []Expression
}

func (e CallExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	resolved := true
	funcName := ""
	var value interface{}
	var info EvaluationInfo

	ref, ok := e.Function.(ReferenceExpr)
	if ok && len(ref.Path) == 1 && ref.Path[0] != "" && ref.Path[0] != "_" {
		funcName = ref.Path[0]
	} else {
		value, info, ok = ResolveExpressionOrPushEvaluation(&e.Function, &resolved, &info, binding)
		if ok {
			_, ok = value.(LambdaValue)
			if !ok {
				debug.Debug("function: no string or lambda value: %T\n", value)
				info.Issue = yaml.NewIssue("function call '%s' requires function name or lambda value", e.Function)
			}
		}
	}

	if !ok {
		debug.Debug("failed to resolve function: %s\n", info.Issue)
		return nil, info, false
	}

	if funcName == "defined" {
		return e.defined(binding)
	}

	values, info, ok := ResolveExpressionListOrPushEvaluation(&e.Arguments, &resolved, nil, binding)

	if !ok {
		debug.Debug("call args failed\n")
		return nil, info, false
	}

	if !resolved {
		return node(e), info, true
	}

	var result yaml.Node
	var sub EvaluationInfo

	switch funcName {
	case "":
		debug.Debug("calling lambda function %#v\n", value)
		result, sub, ok = value.(LambdaValue).Evaluate(values, binding)

	case "static_ips":
		result, sub, ok = func_static_ips(e.Arguments, binding)

	case "join":
		result, sub, ok = func_join(values, binding)

	case "split":
		result, sub, ok = func_split(values, binding)

	case "trim":
		result, sub, ok = func_trim(values, binding)

	case "length":
		result, sub, ok = func_length(values, binding)

	case "exec":
		result, sub, ok = func_exec(values, binding)

	case "min_ip":
		result, sub, ok = func_minIP(values, binding)

	case "max_ip":
		result, sub, ok = func_maxIP(values, binding)

	case "num_ip":
		result, sub, ok = func_numIP(values, binding)

	default:
		info.Issue = yaml.NewIssue("unknown function '%s'", funcName)
		return nil, info, false
	}

	if ok && (result == nil || isExpression(result)) {
		return node(e), sub.Join(info), true
	}
	return result, sub.Join(info), ok
}

func (e CallExpr) String() string {
	args := make([]string, len(e.Arguments))
	for i, e := range e.Arguments {
		args[i] = fmt.Sprintf("%s", e)
	}

	return fmt.Sprintf("%s(%s)", e.Function, strings.Join(args, ", "))
}

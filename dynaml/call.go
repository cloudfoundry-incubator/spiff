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
	
	switch e.Name {
	case "static_ips":
		return func_static_ips(e.Arguments,binding)

	case "join":
		return func_join(values, binding)
		
	case "exec":
		return func_exec(values,binding)
	}
	return nil, info, false
}


func (e CallExpr) String() string {
	args := make([]string, len(e.Arguments))
	for i, e := range e.Arguments {
		args[i] = fmt.Sprintf("%s", e)
	}

	return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))
}


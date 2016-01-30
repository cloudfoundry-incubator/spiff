package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type LambdaExpr struct {
	Names []string
	E     Expression
}

func (e LambdaExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info := DefaultInfo()
	return node(LambdaValue{e, binding.GetLocalBinding()}), info, true
}

func (e LambdaExpr) String() string {
	str := ""
	for _, n := range e.Names {
		str += "," + n
	}
	return fmt.Sprintf("lambda|%s|->%s", str[1:], e.E)
}

type LambdaRefExpr struct {
	Source   Expression
	Path     []string
	StubPath []string
}

func (e LambdaRefExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	var lambda LambdaValue
	resolved := true
	value, info, ok := ResolveExpressionOrPushEvaluation(&e.Source, &resolved, nil, binding)
	if !ok {
		return nil, info, false
	}
	if !resolved {
		return node(e), info, false
	}

	switch v := value.(type) {
	case LambdaValue:
		lambda = v

	case string:
		debug.Debug("LRef: parsing '%s'\n", v)
		expr, err := Parse(v, e.Path, e.StubPath)
		if err != nil {
			debug.Debug("cannot parse: %s\n", err.Error())
			info.Issue = yaml.NewIssue("cannot parse lamba expression '%s'", v)
			return nil, info, false
		}
		lexpr, ok := expr.(LambdaExpr)
		if !ok {
			debug.Debug("no lambda expression: %T\n", expr)
			info.Issue = yaml.NewIssue("'%s' is no lambda expression", v)
			return nil, info, false
		}
		lambda = LambdaValue{lexpr, binding.GetLocalBinding()}

	default:
		info.Issue = yaml.NewIssue("lambda reference must resolve to lambda value or string")
	}
	debug.Debug("found lambda: %s\n", lambda)
	return node(lambda), info, true
}

func (e LambdaRefExpr) String() string {
	return fmt.Sprintf("lambda %s", e.Source)
}

type LambdaValue struct {
	lambda  LambdaExpr
	binding map[string]yaml.Node
}

func (e LambdaValue) String() string {
	binding := ""
	if len(e.binding) > 0 {
		binding = "{"
		for n, v := range e.binding {
			if n != "_" {
				binding += fmt.Sprintf("%s: %v,", n, v.Value())
			}
		}
		binding += "}"
	}
	return fmt.Sprintf("%s%s", binding, e.lambda)
}

func (e LambdaValue) MarshalYAML() (tag string, value interface{}) {
	return "", e.String()
}

func (e LambdaValue) Evaluate(args []interface{}, binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(args) > len(e.lambda.Names) {
		info.Issue = yaml.NewIssue("found %d argument(s), but expects %d", len(args), len(e.lambda.Names))
		return nil, info, false
	}
	inp := map[string]yaml.Node{}
	for n, v := range e.binding {
		inp[n] = v
	}
	debug.Debug("LAMBDA CALL: inherit binding %+v\n", inp)
	inp["_"] = node(e)
	for i, v := range args {
		inp[e.lambda.Names[i]] = node(v)
	}
	debug.Debug("LAMBDA CALL: effective binding %+v\n", inp)

	if len(args) < len(e.lambda.Names) {
		rest := e.lambda.Names[len(args):]
		return node(LambdaValue{LambdaExpr{rest, e.lambda.E}, inp}), DefaultInfo(), true
	}

	return e.lambda.E.Evaluate(newCallBinding(inp, binding))
}

type CallBinding struct {
	Binding
	names map[string]yaml.Node
}

func (c CallBinding) GetLocalBinding() map[string]yaml.Node {
	return c.names
}

func newCallBinding(names map[string]yaml.Node, binding Binding) Binding {
	return CallBinding{binding.WithScope(names), names}
}

/*
func (c MapContext) FindReference(path []string) (yaml.Node, bool) {
	for name, node := range c.names {
		if len(path) >= 1 && path[0] == name {
			debug.Debug("lambda: catch find ref: %v\n", path)
			if len(path) == 1 {
				return node, true
			}
			return yaml.Find(node, path[1:]...)
		}
	}
	debug.Debug("lambda: forward find ref: %v\n", path)
	return c.Binding.FindReference(path)
}
*/

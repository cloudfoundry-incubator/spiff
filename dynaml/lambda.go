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
		if len(v) > 0 && v[0] == '|' {
			v = "lambda " + v
		}
		debug.Debug("LRef: parsing '%s'\n", v)
		expr, err := Parse(v, e.Path, e.StubPath)
		if err != nil {
			debug.Debug("cannot parse: %s\n", err.Error())
			info.Issue = fmt.Sprintf("cannot parse lamba expression '%s'", v)
			return nil, info, false
		}
		lexpr, ok := expr.(LambdaExpr)
		if !ok {
			debug.Debug("no lambda expression: %T\n", expr)
			info.Issue = fmt.Sprintf("'%s' is no lambda expression", v)
			return nil, info, false
		}
		lambda = LambdaValue{lexpr, binding.GetLocalBinding()}

	default:
		info.Issue = "lambda reference must resolve to lambda value or string"
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
	return fmt.Sprintf("%s", e.lambda)
}

func (e LambdaValue) MarshalYAML() (tag string, value interface{}) {
	return "", e.String()
}

func (e LambdaValue) Evaluate(args []interface{}, binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info := DefaultInfo()
	if len(args) != len(e.lambda.Names) {
		info.Issue = fmt.Sprintf("found %d argument(s), but expects %d", len(args), len(e.lambda.Names))
		return nil, info, false
	}
	inp := map[string]yaml.Node{}
	for n, v := range e.binding {
		inp[n] = v
	}
	debug.Debug("LAMBDA CALL: inherit binding %+v\n", inp)
	inp["_"] = node(e)
	for i, name := range e.lambda.Names {
		inp[name] = node(args[i])
	}
	debug.Debug("LAMBDA CALL: effective binding %+v\n", inp)
	ctx := MapContext{binding, inp}
	return e.lambda.E.Evaluate(ctx)
}

type MapContext struct {
	Binding
	names map[string]yaml.Node
}

func (c MapContext) GetLocalBinding() map[string]yaml.Node {
	return c.names
}

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

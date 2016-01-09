package dynaml

import (
	"fmt"
	"sort"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type MapExpr struct {
	A     Expression
	Names []string
	B     Expression
}

func (e MapExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	resolved := true

	debug.Debug("evaluate mapping\n")
	value, info, ok := ResolveExpressionOrPushEvaluation(&e.A, &resolved, nil, binding)
	if !ok {
		return nil, info, false
	}
	if !resolved {
		return node(e), info, ok
	}

	debug.Debug("map: using expression %+v\n", e.B)
	var result []yaml.Node
	switch value.(type) {
	case []yaml.Node:
		result, info, ok = mapList(value.([]yaml.Node), e.Names, e.B, binding)

	case map[string]yaml.Node:
		result, info, ok = mapMap(value.(map[string]yaml.Node), e.Names, e.B, binding)

	default:
		info.Issue = "map or list required for mapping"
		return nil, info, false
	}
	if !ok {
		return nil, info, false
	}
	if result == nil {
		return node(e), info, true
	}
	debug.Debug("map: --> %+v\n", result)
	return node(result), info, true
}

func (e MapExpr) String() string {
	str := ""
	for _, n := range e.Names {
		str = "," + n + str
	}
	return fmt.Sprintf("map[%s|%s|->%s]", e.A, str[1:], e.B)
}

func mapList(source []yaml.Node, names []string, e Expression, binding Binding) ([]yaml.Node, EvaluationInfo, bool) {
	inp := map[string]yaml.Node{}
	result := []yaml.Node{}
	info := DefaultInfo()

	for i, n := range source {
		debug.Debug("map:  mapping for %d: %+v\n", i, n)
		inp[names[0]] = n
		if len(names) > 1 {
			inp[names[1]] = node(i)
		}
		ctx := MapContext{binding, inp}
		mapped, info, ok := e.Evaluate(ctx)
		if !ok {
			debug.Debug("map:  %d %+v: failed\n", i, n)
			return nil, info, false
		}

		_, ok = mapped.Value().(Expression)
		if ok {
			debug.Debug("map:  %d unresolved  -> KEEP\n")
			return nil, info, true
		}
		debug.Debug("map:  %d --> %+v\n", i, mapped)
		result = append(result, mapped)
	}
	return result, info, true
}

func mapMap(source map[string]yaml.Node, names []string, e Expression, binding Binding) ([]yaml.Node, EvaluationInfo, bool) {
	inp := map[string]yaml.Node{}
	result := []yaml.Node{}
	info := DefaultInfo()

	keys := getSortedKeys(source)
	for _, k := range keys {
		n := source[k]
		debug.Debug("map:  mapping for %s: %+v\n", k, n)
		inp[names[0]] = n
		if len(names) > 1 {
			inp[names[1]] = node(k)
		}
		ctx := MapContext{binding, inp}
		mapped, info, ok := e.Evaluate(ctx)
		if !ok {
			debug.Debug("map:  %s %+v: failed\n", k, n)
			return nil, info, false
		}

		_, ok = mapped.Value().(Expression)
		if ok {
			debug.Debug("map:  %d unresolved  -> KEEP\n")
			return nil, info, true
		}
		debug.Debug("map:  %s --> %+v\n", k, mapped)
		result = append(result, mapped)
	}
	return result, info, true
}

func getSortedKeys(unsortedMap map[string]yaml.Node) []string {
	keys := make([]string, len(unsortedMap))
	i := 0
	for k, _ := range unsortedMap {
		keys[i] = k
		i++
	}
	sort.Strings(keys)
	return keys
}

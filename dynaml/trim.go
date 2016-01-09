package dynaml

import (
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func func_trim(arguments []interface{}, binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info := DefaultInfo()
	ok := true

	if len(arguments) > 2 {
		info.Issue = "split takes one or two arguments"
		return nil, info, false
	}

	cutset := " \t"
	if len(arguments) == 2 {
		cutset, ok = arguments[1].(string)
		if !ok {
			info.Issue = "second argument of split must be a string"
			return nil, info, false
		}
	}
	var result interface{}
	switch v := arguments[0].(type) {
	case string:
		result = strings.Trim(v, cutset)
	case []yaml.Node:
		list := make([]yaml.Node, len(v))
		for i, e := range v {
			t, ok := e.Value().(string)
			if !ok {
				info.Issue = "list elements must be strings to be trimmed"
				return nil, info, false
			}
			t = strings.Trim(t, cutset)
			list[i] = node(t)
		}
		result = list
	default:
		info.Issue = "trim accepts only a string or list"
		return nil, info, false
	}

	return node(result), info, true
}

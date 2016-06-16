package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func func_length(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	var result interface{}
	info := DefaultInfo()

	if len(arguments) != 1 {
		info.Issue = yaml.NewIssue("length takes exactly 1 arguments")
		return nil, info, false
	}

	switch v := arguments[0].(type) {
	case []yaml.Node:
		result = len(v)
	case map[string]yaml.Node:
		result = len(v)
	case string:
		result = len(v)
	default:
		info.Issue = yaml.NewIssue("invalid type for function length")
		return nil, info, false

	}
	return result, info, true
}

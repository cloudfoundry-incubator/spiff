package dynaml

import (
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func func_split(arguments []interface{}, binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 2 {
		info.Issue = yaml.NewIssue("split takes exactly 2 arguments")
		return nil, info, false
	}

	sep, ok := arguments[0].(string)
	if !ok {
		info.Issue = yaml.NewIssue("first argument for split must be a string")
		return nil, info, false
	}
	str, ok := arguments[1].(string)
	if !ok {
		info.Issue = yaml.NewIssue("second argument for split must be a string")
		return nil, info, false
	}

	array := strings.Split(str, sep)
	result := make([]yaml.Node, len(array))
	for i, e := range array {
		result[i] = node(e)
	}
	return node(result), info, true
}

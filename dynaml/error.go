package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func func_error(arguments []interface{}, binding Binding) (yaml.Node, EvaluationInfo, bool) {
	n, info, ok := format("error", arguments, binding)
	if !ok {
		return n, info, ok
	}
	info.Issue = yaml.NewIssue("%s", n.Value())
	return nil, info, false
}

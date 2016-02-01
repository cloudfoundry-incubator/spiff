package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func func_eval(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) != 1 {
		info.Issue = yaml.NewIssue("one argument required for 'eval'")
		return nil, info, false
	}

	str, ok := arguments[0].(string)
	if !ok {
		info.Issue = yaml.NewIssue("string argument required for 'eval'")
		return nil, info, false
	}

	expr, err := Parse(str, binding.Path(), binding.StubPath())
	if err != nil {
		info.Issue = yaml.NewIssue("cannot parse expression '%s'", str)
		return nil, info, false
	}
	return expr.Evaluate(binding)
}

package dynaml

import (
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func func_join(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) < 1 {
		return nil, info, false
	}

	args := make([]string, 0)
	for i, arg := range arguments {
		switch v := arg.(type) {
		case string:
			args = append(args, v)
		case int64:
			args = append(args, strconv.FormatInt(v, 10))
		case bool:
			args = append(args, strconv.FormatBool(v))
		case []yaml.Node:
			if i == 0 {
				info.Issue = yaml.NewIssue("first argument for join must be a string")
				return nil, info, false
			}
			for _, elem := range v {
				switch e := elem.Value().(type) {
				case string:
					args = append(args, e)
				case int64:
					args = append(args, strconv.FormatInt(e, 10))
				case bool:
					args = append(args, strconv.FormatBool(e))
				default:
					info.Issue = yaml.NewIssue("elements of list(arg %d) to join must be simple values", i)
					return nil, info, false
				}
			}
		case nil:
		default:
			info.Issue = yaml.NewIssue("argument %d to join must be simple value or list", i)
			return nil, info, false
		}
	}

	return strings.Join(args[1:], args[0]), info, true
}

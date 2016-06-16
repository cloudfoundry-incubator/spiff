package dynaml

import (
	"os"
	"strconv"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var environ []string = os.Environ()

func ReloadEnv() {
	environ = os.Environ()
}

func getenv(name string) (string, bool) {
	name += "="
	for _, s := range environ {
		if strings.HasPrefix(s, name) {
			return s[len(name):], true
		}
	}
	return "", false
}

func func_env(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
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
			info.Issue = yaml.NewIssue("env argument %d must be simple value or list", i)
			return nil, info, false
		}
	}

	if len(args) == 1 {
		s, ok := getenv(args[0])
		if ok {
			return s, info, ok
		} else {
			info.Issue = yaml.NewIssue("environment variable '%s' not set", args[0])
			return nil, info, ok
		}
	} else {
		m := make(map[string]yaml.Node)
		for _, n := range args {
			s, ok := getenv(n)
			if ok {
				m[n] = node(s, nil)
			}
		}
		return m, info, true
	}
}

package dynaml

import (
	"fmt"
	"log"

	"github.com/cloudfoundry-incubator/spiff/yaml"

	"github.com/cloudfoundry-incubator/candiedyaml"
)

func func_format(arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	return format("format", arguments, binding)
}

func format(name string, arguments []interface{}, binding Binding) (interface{}, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(arguments) < 1 {
		info.Issue = yaml.NewIssue("alt least one argument required for '%s'", name)
		return nil, info, false
	}

	args := make([]interface{}, len(arguments))
	for i, arg := range arguments {
		switch v := arg.(type) {
		case []yaml.Node:
			yaml, err := candiedyaml.Marshal(node(v, nil))
			if err != nil {
				log.Fatalln("error marshalling yaml fragment:", err)
			}
			args[i] = string(yaml)
		case map[string]yaml.Node:
			yaml, err := candiedyaml.Marshal(node(v, nil))
			if err != nil {
				log.Fatalln("error marshalling yaml fragment:", err)
			}
			args[i] = string(yaml)
		case TemplateValue:
			yaml, err := candiedyaml.Marshal(v.Orig)
			if err != nil {
				log.Fatalln("error marshalling template:", err)
			}
			args[i] = string(yaml)
		case LambdaValue:
			args[i] = v.String()
		default:
			args[i] = arg
		}
	}

	f, ok := args[0].(string)
	if !ok {
		info.Issue = yaml.NewIssue("%s: format must be string", format)
		return nil, info, false
	}
	return fmt.Sprintf(f, args[1:]...), info, true
}

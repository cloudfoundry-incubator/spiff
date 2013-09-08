package flow

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/vito/spiff/dynaml"
	"github.com/vito/spiff/yaml"
)

var embeddedDynaml = regexp.MustCompile(`^\(\((.*)\)\)$`)

func Flow(source yaml.Node, stubs ...yaml.Node) (yaml.Node, error) {
	result := source

	for {
		next := flow(result, Environment{Stubs: stubs})

		if reflect.DeepEqual(result, next) {
			break
		}

		result = next
	}

	unresolved := findUnresolvedNodes(result)
	if len(unresolved) > 0 {
		return nil, UnresolvedNodes{unresolved}
	}

	return result, nil
}

func flow(root yaml.Node, env Environment) yaml.Node {
	switch root.(type) {
	case map[string]yaml.Node:
		return flowMap(root.(map[string]yaml.Node), env)

	case []yaml.Node:
		return flowList(root.([]yaml.Node), env)

	case string:
		return flowString(root.(string), env)

	case dynaml.Expression:
		result, ok := root.(dynaml.Expression).Evaluate(env)
		if !ok {
			return root
		}

		return result

	default:
		return root
	}
}

func flowMap(root map[string]yaml.Node, env Environment) yaml.Node {
	env = env.WithScope(root)

	newMap := make(map[string]yaml.Node)

	for key, val := range root {
		newMap[key] = flow(val, env.WithPath(key))
	}

	return newMap
}

func flowList(root []yaml.Node, env Environment) yaml.Node {
	newList := []yaml.Node{}

	for idx, val := range root {
		step := stepName(idx, val)
		newList = append(newList, flow(val, env.WithPath(step)))
	}

	return newList
}

func flowString(root string, env Environment) yaml.Node {
	sub := embeddedDynaml.FindStringSubmatch(root)
	if sub == nil {
		return root
	}

	expr, err := dynaml.Parse(sub[1], env.Path)
	if err != nil {
		return root
	}

	return expr
}

func stepName(index int, value yaml.Node) string {
	name, ok := yaml.FindString(value, "name")
	if ok {
		return name
	}

	return fmt.Sprintf("[%d]", index)
}

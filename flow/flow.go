package flow

import (
	"fmt"
	"regexp"

	"github.com/vito/spiff/dynaml"
	"github.com/vito/spiff/yaml"
)

var embeddedDynaml = regexp.MustCompile(`^\(\((.*)\)\)$`)

func Flow(source yaml.Node, stubs ...yaml.Node) (yaml.Node, error) {
	result := source
	didFlow := true

	for didFlow {
		result, didFlow = flow(result, Environment{Stubs: stubs})
	}

	unresolved := findUnresolvedNodes(result)
	if len(unresolved) > 0 {
		return nil, UnresolvedNodes{unresolved}
	}

	return result, nil
}

func flow(root yaml.Node, env Environment) (yaml.Node, bool) {
	switch root.(type) {
	case map[string]yaml.Node:
		return flowMap(root.(map[string]yaml.Node), env)

	case []yaml.Node:
		return flowList(root.([]yaml.Node), env)

	case string:
		return flowString(root.(string), env)

	case dynaml.Expression:
		result := root.(dynaml.Expression).Evaluate(env)
		if result == nil {
			return root, false
		}

		return result, true

	default:
		return root, false
	}
}

func flowMap(root map[string]yaml.Node, env Environment) (yaml.Node, bool) {
	env = env.WithScope(root)

	newMap := make(map[string]yaml.Node)

	flowed := false

	for key, val := range root {
		sub, didFlow := flow(val, env.WithPath(key))
		if didFlow {
			flowed = true
		}

		newMap[key] = sub
	}

	return newMap, flowed
}

func flowList(root []yaml.Node, env Environment) (yaml.Node, bool) {
	newList := []yaml.Node{}

	flowed := false

	for idx, val := range root {
		step := stepName(idx, val)

		sub, didFlow := flow(val, env.WithPath(step))
		if didFlow {
			flowed = true
		}

		newList = append(newList, sub)
	}

	return newList, flowed
}

func flowString(root string, env Environment) (yaml.Node, bool) {
	sub := embeddedDynaml.FindStringSubmatch(root)
	if sub == nil {
		return root, false
	}

	expr, err := dynaml.Parse(sub[1], env.Path)
	if err != nil {
		return root, false
	}

	return expr, true
}

func stepName(index int, value yaml.Node) string {
	name, ok := yaml.FindString(value, "name")
	if ok {
		return name
	}

	return fmt.Sprintf("[%d]", index)
}

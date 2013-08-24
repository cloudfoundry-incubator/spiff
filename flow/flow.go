package flow

import (
	"regexp"

	"github.com/vito/spiff/dynaml"
	"github.com/vito/spiff/yaml"
)

var embeddedDynaml = regexp.MustCompile(`^\(\((.*)\)\)$`)

func Flow(source yaml.Node) yaml.Node {
	result := source
	didFlow := true

	for didFlow {
		result, didFlow = flow(result, Environment{})
	}

	return result
}

func flow(root yaml.Node, env Environment) (yaml.Node, bool) {
	switch root.(type) {
	case map[string]yaml.Node:
		node := root.(map[string]yaml.Node)

		(&env).PushScope(node)

		return flowMap(node, env)

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
	newMap := make(map[string]yaml.Node)

	flowed := false

	for key, val := range root {
		(&env).PushPath(key)

		sub, didFlow := flow(val, env)
		if didFlow {
			flowed = true
		}

		newMap[key] = sub
	}

	return newMap, flowed
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

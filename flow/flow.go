package flow

import (
	"fmt"
	"reflect"
	"regexp"

	"github.com/cloudfoundry-incubator/spiff/dynaml"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var embeddedDynaml = regexp.MustCompile(`^\(\((.*)\)\)$`)

func Flow(source yaml.Node, stubs ...yaml.Node) (yaml.Node, error) {
	result := source

	for {
		next := flow(result, Environment{Stubs: stubs}, true)

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

func flow(root yaml.Node, env Environment, shouldOverride bool) yaml.Node {
	switch root.(type) {
	case map[string]yaml.Node:
		return flowMap(root.(map[string]yaml.Node), env)

	case []yaml.Node:
		return flowList(root.([]yaml.Node), env)

	case dynaml.Expression:
		result, ok := root.(dynaml.Expression).Evaluate(env)
		if !ok {
			return root
		}

		return result
	}

	if shouldOverride {
		overridden, found := env.FindInStubs(env.Path)
		if found {
			return overridden
		}
	}

	str, ok := root.(string)
	if ok {
		return flowString(str, env)
	}

	return root
}

func flowMap(root map[string]yaml.Node, env Environment) yaml.Node {
	env = env.WithScope(root)

	newMap := make(map[string]yaml.Node)

	for key, val := range root {
		if key == "<<" {
			base := flow(val, env, true)
			baseMap, ok := base.(map[string]yaml.Node)
			if ok {
				for k, v := range baseMap {
					newMap[k] = v
				}
			}

			continue
		}

		newMap[key] = flow(val, env.WithPath(key), true)
	}

	return newMap
}

func flowList(root []yaml.Node, env Environment) yaml.Node {
	spliced := []yaml.Node{}

	for _, val := range root {
		subMap, ok := val.(map[string]yaml.Node)
		if ok {
			if len(subMap) == 1 {
				inlineNode, ok := subMap["<<"]
				if ok {
					inline, ok := flow(inlineNode, env, true).([]yaml.Node)
					if ok {
						spliced = append(spliced, inline...)
						continue
					}
				}
			}
		}

		spliced = append(spliced, val)
	}

	newList := []yaml.Node{}

	for idx, val := range spliced {
		step := stepName(idx, val)
		newList = append(newList, flow(val, env.WithPath(step), false))
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

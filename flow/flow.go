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
	if root == nil {
		return root
	}

	switch val := root.Value().(type) {
	case map[string]yaml.Node:
		return flowMap(root, env)

	case []yaml.Node:
		return flowList(root, env)

	case dynaml.Expression:
		result, ok := val.Evaluate(env)
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

	_, ok := root.Value().(string)
	if ok {
		return flowString(root, env)
	}

	return root
}

func flowMap(root yaml.Node, env Environment) yaml.Node {
	rootMap := root.Value().(map[string]yaml.Node)

	env = env.WithScope(rootMap)

	newMap := make(map[string]yaml.Node)

	for key, val := range rootMap {
		if key == "<<" {
			base := flow(val, env, true)
			baseMap, ok := base.Value().(map[string]yaml.Node)
			if ok {
				for k, v := range baseMap {
					newMap[k] = v
				}
			}

			continue
		}

		newMap[key] = flow(val, env.WithPath(key), true)
	}

	return yaml.NewNode(newMap, root.SourceName())
}

func flowList(root yaml.Node, env Environment) yaml.Node {
	rootList := root.Value().([]yaml.Node)

	merged := processMerges(rootList, env)

	newList := []yaml.Node{}

	for idx, val := range merged {
		step := stepName(idx, val)
		newList = append(newList, flow(val, env.WithPath(step), false))
	}

	return yaml.NewNode(newList, root.SourceName())
}

func flowString(root yaml.Node, env Environment) yaml.Node {
	rootString := root.Value().(string)

	sub := embeddedDynaml.FindStringSubmatch(rootString)
	if sub == nil {
		return root
	}

	expr, err := dynaml.Parse(sub[1], env.Path)
	if err != nil {
		return root
	}

	return yaml.NewNode(expr, root.SourceName())
}

func stepName(index int, value yaml.Node) string {
	name, ok := yaml.FindString(value, "name")
	if ok {
		return name
	}

	return fmt.Sprintf("[%d]", index)
}

func processMerges(root []yaml.Node, env Environment) []yaml.Node {
	spliced := []yaml.Node{}

	for _, val := range root {
		if val == nil {
			continue
		}

		subMap, ok := val.Value().(map[string]yaml.Node)
		if ok {
			if len(subMap) == 1 {
				inlineNode, ok := subMap["<<"]
				if ok {
					inline, ok := flow(inlineNode, env, true).Value().([]yaml.Node)

					if ok {
						inlineNew := newEntries(inline, root)
						spliced = append(spliced, inlineNew...)
						continue
					}
				}
			}
		}

		spliced = append(spliced, val)
	}

	return spliced
}

func newEntries(a []yaml.Node, b []yaml.Node) []yaml.Node {
	added := []yaml.Node{}

	for _, val := range a {
		name, ok := yaml.FindString(val, "name")
		if ok {
			_, found := yaml.Find(yaml.NewNode(b, "some map"), name) // TODO
			if found {
				continue
			}
		}

		added = append(added, val)
	}

	return added
}

package flow

import (
	"regexp"
	"strconv"

	"github.com/vito/spiff/yaml"
)

type Scope []map[string]yaml.Node

type Environment struct {
	Scope Scope
	Path  []string

	Stubs []yaml.Node
}

func (e Environment) FindFromRoot(path []string) yaml.Node {
	return findInPath(path, e.Scope[0])
}

func (e Environment) FindReference(path []string) yaml.Node {
	root, found := resolveSymbol(path[0], e.Scope)
	if !found {
		return nil
	}

	return findInPath(path[1:], root)
}

func (e Environment) FindInStubs(path []string) yaml.Node {
	for _, stub := range e.Stubs {
		found := findInPath(path, stub)
		if found != nil {
			return found
		}
	}

	return nil
}

func (e Environment) WithScope(step map[string]yaml.Node) Environment {
	e.Scope = append(e.Scope, step)
	return e
}

func (e Environment) WithPath(step string) Environment {
	e.Path = append(e.Path, step)
	return e
}

func findInPath(path []string, root yaml.Node) yaml.Node {
	here := root

	for _, step := range path {
		if here == nil {
			return nil
		}

		var found bool

		here, found = nextStep(step, here)
		if !found {
			return nil
		}
	}

	return here
}

func nextStep(step string, here yaml.Node) (yaml.Node, bool) {
	found := false

	switch here.(type) {
	case map[string]yaml.Node:
		here, found = here.(map[string]yaml.Node)[step]
	case []yaml.Node:
		here, found = stepThroughList(here.([]yaml.Node), step)
	default:
	}

	return here, found
}

var listIndex = regexp.MustCompile(`^\[(\d+)\]$`)

func stepThroughList(here []yaml.Node, step string) (yaml.Node, bool) {
	match := listIndex.FindStringSubmatch(step)
	if match != nil {
		index, err := strconv.Atoi(match[1])
		if err != nil {
			panic(err)
		}

		if len(here) <= index {
			return nil, false
		}

		return here[index], true
	}

	for _, sub := range here {
		subMap, ok := sub.(map[string]yaml.Node)
		if !ok {
			continue
		}

		name, ok := subMap["name"]
		if !ok {
			continue
		}

		nameString, ok := name.(string)
		if !ok {
			continue
		}

		if nameString == step {
			return subMap, true
		}
	}

	return here, false
}

func resolveSymbol(name string, context Scope) (yaml.Node, bool) {
	for i := len(context); i > 0; i-- {
		ctx := context[i-1]
		val := ctx[name]
		if val != nil {
			return val, true
		}
	}

	return nil, false
}

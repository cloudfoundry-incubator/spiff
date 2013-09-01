package yaml

import (
	"regexp"
	"strconv"
)

var listIndex = regexp.MustCompile(`^\[(\d+)\]$`)

func Find(root Node, path ...string) (Node, bool) {
	here := root

	for _, step := range path {
		if here == nil {
			return nil, false
		}

		var found bool

		here, found = nextStep(step, here)
		if !found {
			return nil, false
		}
	}

	return here, true
}

func FindString(root Node, path ...string) (string, bool) {
	node, ok := Find(root, path...)
	if !ok {
		return "", false
	}

	val, ok := node.(string)
	return val, ok
}

func FindInt(root Node, path ...string) (int, bool) {
	node, ok := Find(root, path...)
	if !ok {
		return 0, false
	}

	val, ok := node.(int)
	return val, ok
}

func nextStep(step string, here Node) (Node, bool) {
	found := false

	switch here.(type) {
	case map[string]Node:
		here, found = here.(map[string]Node)[step]
	case []Node:
		here, found = stepThroughList(here.([]Node), step)
	default:
	}

	return here, found
}

func stepThroughList(here []Node, step string) (Node, bool) {
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
		subMap, ok := sub.(map[string]Node)
		if !ok {
			continue
		}

		name, ok := FindString(sub, "name")
		if !ok {
			continue
		}

		if name == step {
			return subMap, true
		}
	}

	return here, false
}

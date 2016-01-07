package yaml

import (
	"regexp"
	"strconv"
	"strings"
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

	val, ok := node.Value().(string)
	return val, ok
}

func FindInt(root Node, path ...string) (int64, bool) {
	node, ok := Find(root, path...)
	if !ok {
		return 0, false
	}

	val, ok := node.Value().(int64)
	return val, ok
}

func nextStep(step string, here Node) (Node, bool) {
	found := false

	switch v := here.Value().(type) {
	case map[string]Node:
		here, found = v[step]
	case []Node:
		here, found = stepThroughList(v, step, here.KeyName())
	default:
	}

	return here, found
}

func stepThroughList(here []Node, step string, key string) (Node, bool) {
	match := listIndex.FindStringSubmatch(step)
	if match != nil {
		index, err := strconv.Atoi(match[1])
		if err != nil {
			panic(err)
		}

		if len(here) <= index {
			return nil, false
		}

		for i := 0; i <= index; i++ {
			_, ok := UnresolvedMerge(here[i])
			if ok {
				return nil, false
			}
		}
		return here[index], true
	}

	if key == "" {
		key = "name"
	}
	split := strings.Index(step, ":")
	if split > 0 {
		key = step[:split]
		step = step[split+1:]
	}

	for _, sub := range here {
		_, ok := sub.Value().(map[string]Node)
		if !ok {
			continue
		}

		name, ok := FindString(sub, key)
		if !ok {
			continue
		}

		if name == step {
			return sub, true
		}
	}

	return nil, false
}

func PathComponent(step string) string {
	split := strings.Index(step, ":")
	if split > 0 {
		return step[split+1:]
	}
	return step
}

func UnresolvedMerge(node Node) (Node, bool) {
	subMap, ok := node.Value().(map[string]Node)
	if ok {
		if len(subMap) == 1 {
			inlineNode, ok := subMap["<<"]
			if ok {
				return inlineNode, true
			}
		}
	}
	return nil, false
}

package compare

import (
	"fmt"
	"reflect"

	"github.com/vito/spiff/yaml"
)

type Diff struct {
	A yaml.Node
	B yaml.Node

	Path []string
}

func Compare(a, b yaml.Node) []Diff {
	return compare(a, b, []string{})
}

func compare(a, b yaml.Node, path []string) []Diff {
	mismatch := Diff{A: a, B: b, Path: path}

	switch a.(type) {
	case map[string]yaml.Node:
		switch b.(type) {
		case map[string]yaml.Node:
			return compareMap(a.(map[string]yaml.Node), b.(map[string]yaml.Node), path)

		case []yaml.Node:
			toMap := listToMap(b.([]yaml.Node))

			if toMap != nil {
				return compareMap(a.(map[string]yaml.Node), toMap, path)
			} else {
				return []Diff{mismatch}
			}

		default:
			return []Diff{mismatch}
		}

	case []yaml.Node:
		return compareList(a.([]yaml.Node), b.([]yaml.Node), path)

	default:
		atype := reflect.TypeOf(a)
		btype := reflect.TypeOf(b)

		if atype != btype {
			return []Diff{mismatch}
		}

		if a != b {
			return []Diff{Diff{A: a, B: b, Path: path}}
		}
	}

	return []Diff{}
}

func listToMap(list []yaml.Node) map[string]yaml.Node {
	toMap := make(map[string]yaml.Node)

	for _, val := range list {
		asMap, ok := val.(map[string]yaml.Node)
		if !ok {
			return nil
		}

		nameNode, ok := asMap["name"]
		if !ok {
			return nil
		}

		name, ok := nameNode.(string)
		if !ok {
			return nil
		}

		newMap := make(map[string]yaml.Node)
		for key, val := range asMap {
			if key != "name" {
				newMap[key] = val
			}
		}

		toMap[name] = newMap
	}

	return toMap
}

func compareMap(a, b map[string]yaml.Node, path []string) []Diff {
	diff := []Diff{}

	for key, aval := range a {
		bval, present := b[key]
		if present {
			diff = append(diff, compare(aval, bval, addPath(path, key))...)
		} else {
			diff = append(diff, Diff{A: aval, B: nil, Path: addPath(path, key)})
		}
	}

	for key, bval := range b {
		_, present := a[key]
		if !present {
			diff = append(diff, Diff{A: nil, B: bval, Path: addPath(path, key)})
			continue
		}
	}

	return diff
}

func compareList(a, b []yaml.Node, path []string) []Diff {
	diff := []Diff{}

	if len(path) == 1 && path[0] == "jobs" {
		return compareJobs(a, b, path)
	}

	for index, aval := range a {
		key, bval, found := findByNameOrIndex(aval, b, index)

		if !found {
			diff = append(diff, Diff{A: aval, B: nil, Path: addPath(path, key)})
			continue
		}

		diff = append(diff, compare(aval, bval, addPath(path, key))...)
	}

	for index, bval := range b {
		key := fmt.Sprintf("[%d]", index)

		if len(a) <= index {
			diff = append(diff, Diff{A: nil, B: bval, Path: addPath(path, key)})
			continue
		}
	}

	return diff
}

func compareJobs(ajobs, bjobs []yaml.Node, path []string) []Diff {
	diff := []Diff{}

	for index, ajob := range ajobs {
		key := fmt.Sprintf("[%d]", index)

		amap, ok := ajob.(map[string]yaml.Node)
		if !ok {
			panic("non-map job")
		}

		bjob := bjobs[index]
		bmap, ok := bjob.(map[string]yaml.Node)
		if !ok {
			panic("non-map job")
		}

		aNameNode, ok := amap["name"]
		if !ok {
			panic("job without name")
		}

		bNameNode, ok := bmap["name"]
		if !ok {
			panic("job without name")
		}

		aname, ok := aNameNode.(string)
		if !ok {
			panic("job with non-string name")
		}

		bname, ok := bNameNode.(string)
		if !ok {
			panic("job with non-string name")
		}

		if aname != bname {
			diff = append(diff, Diff{A: aname, B: bname, Path: addPath(path, key, "name")})
			continue
		}

		diff = append(diff, compare(ajob, bjob, addPath(path, aname))...)
	}

	return diff
}

func findByNameOrIndex(node yaml.Node, others []yaml.Node, index int) (string, yaml.Node, bool) {
	nodeMap, ok := node.(map[string]yaml.Node)
	if !ok {
		return findByIndex(others, index)
	}

	nameNode, hasName := nodeMap["name"]
	if !hasName {
		return findByIndex(others, index)
	}

	name, ok := nameNode.(string)
	if !ok {
		return findByIndex(others, index)
	}

	key, node, found := findByName(name, others)
	if !found {
		return findByIndex(others, index)
	}

	return key, node, true
}

func findByName(name string, nodes []yaml.Node) (string, yaml.Node, bool) {
	for _, node := range nodes {
		nodeMap, ok := node.(map[string]yaml.Node)
		if !ok {
			continue
		}

		nameNode, hasName := nodeMap["name"]
		if !hasName {
			continue
		}

		otherName, ok := nameNode.(string)
		if !ok {
			continue
		}

		if otherName == name {
			return name, node, true
		}
	}

	return "", nil, false
}

func findByIndex(nodes []yaml.Node, index int) (string, yaml.Node, bool) {
	key := fmt.Sprintf("[%d]", index)

	if len(nodes) <= index {
		return key, nil, false
	}

	return key, nodes[index], true
}

// cannot use straight append for this, as it will overwrite
// previous steps, since it reuses the slice
//
// e.g. with inital path A:
//    append(A, "a")
//    append(A, "b")
//
// will result in all previous A/a paths becoming A/b
func addPath(path []string, steps ...string) []string {
	newPath := make([]string, len(path))
	copy(newPath, path)
	return append(newPath, steps...)
}

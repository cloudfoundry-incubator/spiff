package flow

import (
	"fmt"
	"reflect"

	"github.com/vito/spiff/dynaml"
	"github.com/vito/spiff/yaml"
)

type resolvedNode struct {
	Value interface{}
}

type UnresolvedNodes struct {
	Nodes []dynaml.Expression
}

func (e UnresolvedNodes) Error() string {
	message := "unresolved nodes:"

	for _, expr := range e.Nodes {
		message = fmt.Sprintf("%s\n\t%T%v", message, expr, expr)
	}

	return message
}

func ResolveNodes(source yaml.Node) (result yaml.Node, unresolved []dynaml.Expression) {
	result = source

	var next yaml.Node

	for {
		next, unresolved = resolveNodesOneStep(result)

		if reflect.DeepEqual(result, next) {
			break
		}

		result = next
	}

	return result, unresolved
}

func resolveNodesOneStep(root yaml.Node) (result yaml.Node, nodes []dynaml.Expression) {
	switch root.(type) {
	case map[string]yaml.Node:
		newMap := map[string]yaml.Node{}

		for key, val := range root.(map[string]yaml.Node) {
			res, unresolved := resolveNodesOneStep(val)
			newMap[key] = res
			nodes = append(nodes, unresolved...)
		}

		result = newMap

	case []yaml.Node:
		newList := []yaml.Node{}

		for _, val := range root.([]yaml.Node) {
			res, unresolved := resolveNodesOneStep(val)
			newList = append(newList, res)
			nodes = append(nodes, unresolved...)
		}

		result = newList

	case dynaml.Expression:
		result = root
		nodes = append(nodes, root.(dynaml.Expression))

	case resolvedNode:
		result = root.(resolvedNode).Value

	default:
		result = root
	}

	return result, nodes
}

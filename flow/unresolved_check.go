package flow

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/dynaml"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type UnresolvedNodes struct {
	Nodes []yaml.Node
}

func (e UnresolvedNodes) Error() string {
	message := "unresolved nodes:"

	for _, node := range e.Nodes {
		message = fmt.Sprintf(
			"%s\n\t(( %s )) in %s",
			message,
			node.Value(),
			node.SourceName(),
		)
	}

	return message
}

func findUnresolvedNodes(root yaml.Node) (nodes []yaml.Node) {
	if root == nil {
		return nodes
	}

	switch val := root.Value().(type) {
	case map[string]yaml.Node:
		for _, val := range val {
			nodes = append(nodes, findUnresolvedNodes(val)...)
		}

	case []yaml.Node:
		for _, val := range val {
			nodes = append(nodes, findUnresolvedNodes(val)...)
		}

	case dynaml.Expression:
		nodes = append(nodes, root)
	}

	return nodes
}

package flow

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/dynaml"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type UnresolvedNodes struct {
	Nodes []dynaml.Expression
}

func (e UnresolvedNodes) Error() string {
	message := "unresolved nodes:"

	for _, expr := range e.Nodes {
		message = fmt.Sprintf("%s\n\t(( %s ))", message, expr)
	}

	return message
}

func findUnresolvedNodes(root yaml.Node) (nodes []dynaml.Expression) {
	switch root.(type) {
	case map[string]yaml.Node:
		for _, val := range root.(map[string]yaml.Node) {
			nodes = append(nodes, findUnresolvedNodes(val)...)
		}

	case []yaml.Node:
		for _, val := range root.([]yaml.Node) {
			nodes = append(nodes, findUnresolvedNodes(val)...)
		}

	case dynaml.Expression:
		nodes = append(nodes, root.(dynaml.Expression))
	}

	return nodes
}

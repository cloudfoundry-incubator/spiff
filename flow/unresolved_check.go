package flow

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/dynaml"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type UnresolvedNodes struct {
	Nodes []UnresolvedNode
}

type UnresolvedNode struct {
	yaml.Node

	Context []string
	Path    []string
}

func (e UnresolvedNodes) Error() string {
	message := "unresolved nodes:"

	for _, node := range e.Nodes {
		message = fmt.Sprintf(
			"%s\n\t(( %s ))\tin %s\t%s\t(%s)",
			message,
			node.Value(),
			node.SourceName(),
			strings.Join(node.Context, "."),
			strings.Join(node.Path, "."),
		)
	}

	return message
}

func findUnresolvedNodes(root yaml.Node, context ...string) (nodes []UnresolvedNode) {
	if root == nil {
		return nodes
	}

	switch val := root.Value().(type) {
	case map[string]yaml.Node:
		for key, val := range val {
			nodes = append(
				nodes,
				findUnresolvedNodes(val, addContext(context, key)...)...,
			)
		}

	case []yaml.Node:
		for i, val := range val {
			context := addContext(context, fmt.Sprintf("[%d]", i))

			nodes = append(
				nodes,
				findUnresolvedNodes(val, context...)...,
			)
		}

	case dynaml.Expression:
		var path []string
		switch val := root.Value().(type) {
		case dynaml.AutoExpr:
			path = val.Path
		case dynaml.MergeExpr:
			path = val.Path
		}

		nodes = append(nodes, UnresolvedNode{
			Node:    root,
			Context: context,
			Path:    path,
		})
	}

	return nodes
}

func addContext(context []string, step string) []string {
	dup := make([]string, len(context))
	copy(dup, context)
	return append(dup, step)
}

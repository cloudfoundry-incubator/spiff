package dynaml

import (
	"fmt"
	"strings"

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
	format := ""

	for _, node := range e.Nodes {
		issue := node.Issue()
		if issue != "" {
			issue = "\t" + issue
		}
		switch node.Value().(type) {
		case Expression:
			format = "%s\n\t(( %s ))\tin %s\t%s\t(%s)%s"
		default:
			format = "%s\n\t%s\tin %s\t%s\t(%s)%s"
		}
		message = fmt.Sprintf(
			format,
			message,
			node.Value(),
			node.SourceName(),
			strings.Join(node.Context, "."),
			strings.Join(node.Path, "."),
			issue,
		)
	}

	return message
}

func FindUnresolvedNodes(root yaml.Node, context ...string) (nodes []UnresolvedNode) {
	if root == nil {
		return nodes
	}

	switch val := root.Value().(type) {
	case map[string]yaml.Node:
		for key, val := range val {
			nodes = append(
				nodes,
				FindUnresolvedNodes(val, addContext(context, key)...)...,
			)
		}

	case []yaml.Node:
		for i, val := range val {
			context := addContext(context, fmt.Sprintf("[%d]", i))

			nodes = append(
				nodes,
				FindUnresolvedNodes(val, context...)...,
			)
		}

	case Expression:
		var path []string
		switch val := root.Value().(type) {
		case AutoExpr:
			path = val.Path
		case MergeExpr:
			path = val.Path
		}

		nodes = append(nodes, UnresolvedNode{
			Node:    root,
			Context: context,
			Path:    path,
		})

	case string:
		if yaml.EmbeddedDynaml(root) != nil {
			nodes = append(nodes, UnresolvedNode{
				Node:    yaml.IssueNode(root, "unparseable expression"),
				Context: context,
				Path:    []string{},
			})
		}
	}

	return nodes
}

func addContext(context []string, step string) []string {
	dup := make([]string, len(context))
	copy(dup, context)
	return append(dup, step)
}

func isResolved(node yaml.Node) bool {
	if node == nil {
		return true
	}
	switch node.Value().(type) {
	case Expression:
		return false
	case []yaml.Node:
		for _, n := range node.Value().([]yaml.Node) {
			if !isResolved(n) {
				return false
			}
		}
		return true
	case map[string]yaml.Node:
		for _, n := range node.Value().(map[string]yaml.Node) {
			if !isResolved(n) {
				return false
			}
		}
		return true

	case string:
		if yaml.EmbeddedDynaml(node) != nil {
			return false
		}
		return true
	default:
		return true
	}
}

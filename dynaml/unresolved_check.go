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

func (e UnresolvedNodes) Issue(message string) yaml.Issue {
	result := yaml.NewIssue(message)
	format := ""

	for _, node := range e.Nodes {
		issue := node.Issue()
		msg := issue.Issue
		if msg != "" {
			msg = "\t" + msg
		}
		switch node.Value().(type) {
		case Expression:
			format = "\t(( %s ))\tin %s\t%s\t(%s)%s"
		default:
			format = "\t%s\tin %s\t%s\t(%s)%s"
		}
		message = fmt.Sprintf(
			format,
			node.Value(),
			node.SourceName(),
			strings.Join(node.Context, "."),
			strings.Join(node.Path, "."),
			msg,
		)
		issue.Issue = message
		result.Nested = append(result.Nested, issue)
	}
	return result
}

func (e UnresolvedNodes) Error() string {
	message := "unresolved nodes:"
	format := ""

	for _, node := range e.Nodes {
		issue := node.Issue()
		msg := issue.Issue
		if msg != "" {
			msg = "\t" + msg
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
			msg,
		)
		message += nestedIssues("\t", issue)
	}

	return message
}

func nestedIssues(gap string, issue yaml.Issue) string {
	message := ""
	if issue.Nested != nil {
		for _, sub := range issue.Nested {
			message = message + "\n" + gap + sub.Issue
			message += nestedIssues(gap+"\t", sub)
		}
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
				Node:    yaml.IssueNode(root, yaml.Issue{Issue: "unparseable expression"}),
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

func isExpression(val interface{}) bool {
	if val == nil {
		return false
	}
	_, ok := val.(Expression)
	return ok
}

func isLocallyResolved(node yaml.Node) bool {
	switch v := node.Value().(type) {
	case Expression:
		return false
	case map[string]yaml.Node:
		if !yaml.IsMapResolved(v) {
			return false
		}
	case []yaml.Node:
		if !yaml.IsListResolved(v) {
			return false
		}
	default:
	}

	return true
}

func isResolved(node yaml.Node) bool {
	return node == nil || isResolvedValue(node.Value())
}

func isResolvedValue(val interface{}) bool {
	if val == nil {
		return true
	}
	switch v := val.(type) {
	case Expression:
		return false
	case []yaml.Node:
		for _, n := range v {
			if !isResolved(n) {
				return false
			}
		}
		return true
	case map[string]yaml.Node:
		for _, n := range v {
			if !isResolved(n) {
				return false
			}
		}
		return true

	case string:
		if yaml.EmbeddedDynaml(node(val, nil)) != nil {
			return false
		}
		return true
	default:
		return true
	}
}

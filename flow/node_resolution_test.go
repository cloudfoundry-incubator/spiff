package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/dynaml"
	"github.com/vito/spiff/yaml"
)

var _ = Describe("Resolving nodes", func() {
	It("replaces resolved nodes with their values", func() {
		result := map[string]yaml.Node{
			"foo": resolvedNode{"bar"},
		}

		expected := map[string]yaml.Node{
			"foo": "bar",
		}

		resolved, _ := ResolveNodes(result)

		Expect(resolved).To(Equal(expected))
	})

	It("formats a message listing the nodes", func() {
		err := UnresolvedNodes{
			Nodes: []dynaml.Expression{
				dynaml.AutoExpr{
					Path: []string{"foo", "bar"},
				},
				dynaml.MergeExpr{
					Path: []string{"fizz"},
				},
			},
		}

		Expect(err.Error()).To(Equal(
			`unresolved nodes:
	dynaml.AutoExpr{[foo bar]}
	dynaml.MergeExpr{[fizz]}`))
	})
})

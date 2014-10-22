package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/shutej/spiff/dynaml"
	"github.com/shutej/spiff/yaml"
)

var _ = Describe("Reporting unresolved nodes", func() {
	It("formats a message listing the nodes", func() {
		err := UnresolvedNodes{
			Nodes: []UnresolvedNode{
				{
					Node: yaml.NewNode(
						dynaml.MergeExpr{},
						"some-other-file.yml",
					),
					Context: []string{"fizz", "[2]", "buzz"},
				},
			},
		}

		Expect(err.Error()).To(Equal(
			`unresolved nodes:
	(( merge ))	in some-other-file.yml	fizz.[2].buzz`))
	})
})

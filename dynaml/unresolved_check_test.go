package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var _ = Describe("Reporting unresolved nodes", func() {
	It("formats a message listing the nodes", func() {
		err := UnresolvedNodes{
			Nodes: []UnresolvedNode{
				{
					Node: yaml.NewNode(
						AutoExpr{},
						"some-file.yml",
					),
					Context: []string{"foo", "bar"},
					Path:    []string{"foo", "bar"},
				},
				{
					Node: yaml.NewNode(
						MergeExpr{},
						"some-other-file.yml",
					),
					Context: []string{"fizz", "[2]", "buzz"},
					Path:    []string{"fizz", "fizzbuzz", "buzz"},
				},
			},
		}

		Expect(err.Error()).To(Equal(
			`unresolved nodes:
	(( auto ))	in some-file.yml	foo.bar	(foo.bar)
	(( merge ))	in some-other-file.yml	fizz.[2].buzz	(fizz.fizzbuzz.buzz)`))
	})
})

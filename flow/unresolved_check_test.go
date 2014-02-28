package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/dynaml"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var _ = Describe("Reporting unresolved nodes", func() {
	It("formats a message listing the nodes", func() {
		err := UnresolvedNodes{
			Nodes: []yaml.Node{
				yaml.NewNode(
					dynaml.AutoExpr{
						Path: []string{"foo", "bar"},
					},
					"some-file.yml",
				),
				yaml.NewNode(
					dynaml.MergeExpr{
						Path: []string{"fizz"},
					},
					"some-other-file.yml",
				),
			},
		}

		Expect(err.Error()).To(Equal(
			`unresolved nodes:
	(( auto )) in some-file.yml
	(( merge )) in some-other-file.yml`))
	})
})

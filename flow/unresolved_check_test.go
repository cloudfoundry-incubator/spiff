package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/dynaml"
)

var _ = Describe("Reporting unresolved nodes", func() {
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

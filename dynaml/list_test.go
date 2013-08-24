package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/yaml"
)

var _ = d.Describe("lists", func() {
	d.It("evaluates to an array of nodes", func() {
		expr := ListExpr{
			[]Expression{
				IntegerExpr{1},
				StringExpr{"two"},
			},
		}
		Expect(expr.Evaluate(FakeContext{})).To(Equal([]yaml.Node{1, "two"}))
	})
})

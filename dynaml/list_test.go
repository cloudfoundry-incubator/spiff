package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("lists", func() {
	d.It("evaluates to an array of nodes", func() {
		expr := ListExpr{
			[]Expression{
				IntegerExpr{1},
				StringExpr{"two"},
			},
		}
		Expect(expr.Evaluate(FakeContext{})).To(Equal([]Node{1, "two"}))
	})
})

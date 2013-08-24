package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("addition", func() {
	d.It("adds both numbers", func() {
		expr := AdditionExpr{
			IntegerExpr{2},
			IntegerExpr{3},
		}

		Expect(expr.Evaluate(FakeContext{})).To(Equal(5))
	})

	d.Context("when the left-hand side is not an integer", func() {
		d.It("returns nil", func() {
			expr := AdditionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr.Evaluate(FakeContext{})).To(BeNil())
		})
	})

	d.Context("when the right-hand side is not an integer", func() {
		d.It("returns nil", func() {
			expr := AdditionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr.Evaluate(FakeContext{})).To(BeNil())
		})
	})
})

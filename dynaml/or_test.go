package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("or", func() {
	d.Context("when both sides are nil", func() {
		expr := OrExpr{
			ReferenceExpr{},
			ReferenceExpr{},
		}

		Expect(expr.Evaluate(FakeContext{})).To(BeNil())
	})

	d.Context("when the left-hand side is nil", func() {
		d.It("returns the right-hand side", func() {
			expr := OrExpr{
				ReferenceExpr{},
				IntegerExpr{2},
			}

			Expect(expr.Evaluate(FakeContext{})).To(Equal(2))
		})
	})

	d.Context("when the right-hand side is nil", func() {
		d.It("returns the left-hand side", func() {
			expr := OrExpr{
				IntegerExpr{1},
				ReferenceExpr{},
			}

			Expect(expr.Evaluate(FakeContext{})).To(Equal(1))
		})
	})
})

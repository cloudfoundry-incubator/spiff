package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("or", func() {
	d.Context("when both sides fail", func() {
		expr := OrExpr{
			ReferenceExpr{},
			ReferenceExpr{},
		}

		Expect(expr).To(FailToEvaluate(FakeContext{}))
	})

	d.Context("when the left-hand side fails", func() {
		d.It("returns the right-hand side", func() {
			expr := OrExpr{
				ReferenceExpr{},
				IntegerExpr{2},
			}

			Expect(expr).To(EvaluateAs(2, FakeContext{}))
		})
	})

	d.Context("when the right-hand side fails", func() {
		d.It("returns the left-hand side", func() {
			expr := OrExpr{
				IntegerExpr{1},
				ReferenceExpr{},
			}

			Expect(expr).To(EvaluateAs(1, FakeContext{}))
		})
	})
})

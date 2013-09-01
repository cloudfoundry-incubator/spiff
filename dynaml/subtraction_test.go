package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("subtraction", func() {
	d.It("subtracts both numbers", func() {
		expr := SubtractionExpr{
			IntegerExpr{7},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs(4, FakeContext{}))
	})

	d.Context("when the left-hand side is not an integer", func() {
		d.It("fails", func() {
			expr := SubtractionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeContext{}))
		})
	})

	d.Context("when the right-hand side is not an integer", func() {
		d.It("fails", func() {
			expr := SubtractionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeContext{}))
		})
	})
})

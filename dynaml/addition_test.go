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

		Expect(expr).To(EvaluateAs(5, FakeContext{}))
	})

	d.Context("when the left-hand side is not an integer", func() {
		d.It("fails", func() {
			expr := AdditionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeContext{}))
		})
	})

	d.Context("when the right-hand side is not an integer", func() {
		d.It("fails", func() {
			expr := AdditionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeContext{}))
		})
	})
})

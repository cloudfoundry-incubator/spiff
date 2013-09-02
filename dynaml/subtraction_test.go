package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("subtraction", func() {
	It("subtracts both numbers", func() {
		expr := SubtractionExpr{
			IntegerExpr{7},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs(4, FakeBinding{}))
	})

	Context("when the left-hand side is not an integer", func() {
		It("fails", func() {
			expr := SubtractionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is not an integer", func() {
		It("fails", func() {
			expr := SubtractionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
})

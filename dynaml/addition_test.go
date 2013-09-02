package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("addition", func() {
	It("adds both numbers", func() {
		expr := AdditionExpr{
			IntegerExpr{2},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs(5, FakeBinding{}))
	})

	Context("when the left-hand side is not an integer", func() {
		It("fails", func() {
			expr := AdditionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is not an integer", func() {
		It("fails", func() {
			expr := AdditionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
})

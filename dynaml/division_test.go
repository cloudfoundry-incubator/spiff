package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("division", func() {
	It("divides both numbers", func() {
		expr := DivisionExpr{
			IntegerExpr{6},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs(2, FakeBinding{}))
	})

	Context("when the left-hand side is not an integer", func() {
		It("fails", func() {
			expr := DivisionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is not an integer", func() {
		It("fails", func() {
			expr := DivisionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is zero", func() {
		It("fails", func() {
			expr := DivisionExpr{
				IntegerExpr{2},
				IntegerExpr{0},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
})

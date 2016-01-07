package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("multiplication", func() {
	It("multiplies both numbers", func() {
		expr := MultiplicationExpr{
			IntegerExpr{2},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs(6, FakeBinding{}))
	})

	Context("when the left-hand side is not an integer", func() {
		It("fails", func() {
			expr := MultiplicationExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is not an integer", func() {
		It("fails", func() {
			expr := MultiplicationExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
})

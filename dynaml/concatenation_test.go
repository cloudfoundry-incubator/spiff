package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("concatenation", func() {
	It("concatenates two strings", func() {
		expr := ConcatenationExpr{
			StringExpr{"one"},
			StringExpr{"two"},
		}

		Expect(expr).To(EvaluateAs("onetwo", FakeBinding{}))
	})

	Context("when the left-hand side is not a string", func() {
		It("fails", func() {
			expr := ConcatenationExpr{
				StringExpr{"one"},
				IntegerExpr{42},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is not a string", func() {
		It("fails", func() {
			expr := ConcatenationExpr{
				IntegerExpr{42},
				StringExpr{"two"},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
})

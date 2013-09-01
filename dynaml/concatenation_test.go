package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("concatenation", func() {
	d.It("concatenates two strings", func() {
		expr := ConcatenationExpr{
			StringExpr{"one"},
			StringExpr{"two"},
		}

		Expect(expr).To(EvaluateAs("onetwo", FakeContext{}))
	})

	d.Context("when the left-hand side is not a string", func() {
		d.It("fails", func() {
			expr := ConcatenationExpr{
				StringExpr{"one"},
				IntegerExpr{42},
			}

			Expect(expr).To(FailToEvaluate(FakeContext{}))
		})
	})

	d.Context("when the right-hand side is not a string", func() {
		d.It("fails", func() {
			expr := ConcatenationExpr{
				IntegerExpr{42},
				StringExpr{"two"},
			}

			Expect(expr).To(FailToEvaluate(FakeContext{}))
		})
	})
})

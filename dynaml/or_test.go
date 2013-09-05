package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("or", func() {
	Context("when both sides fail", func() {
		It("fails", func() {
			expr := OrExpr{
				FailingExpr{},
				FailingExpr{},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the left-hand side fails", func() {
		It("returns the right-hand side", func() {
			expr := OrExpr{
				FailingExpr{},
				IntegerExpr{2},
			}

			Expect(expr).To(EvaluateAs(2, FakeBinding{}))
		})
	})

	Context("when the right-hand side fails", func() {
		It("returns the left-hand side", func() {
			expr := OrExpr{
				IntegerExpr{1},
				FailingExpr{},
			}

			Expect(expr).To(EvaluateAs(1, FakeBinding{}))
		})
	})

	Context("when the left-hand side is nil", func() {
		It("returns the left-hand side", func() {
			expr := OrExpr{
				NilExpr{},
				FailingExpr{},
			}

			Expect(expr).To(EvaluateAs(nil, FakeBinding{}))
		})
	})

	Context("when the right side is nil and the left fails", func() {
		It("returns the left-hand side", func() {
			expr := OrExpr{
				FailingExpr{},
				NilExpr{},
			}

			Expect(expr).To(EvaluateAs(nil, FakeBinding{}))
		})
	})
})

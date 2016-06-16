package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("logical or", func() {
	Context("on boolean", func() {
		It("true or true", func() {
			expr := LogOrExpr{
				BooleanExpr{true},
				BooleanExpr{true},
			}

			Expect(expr).To(EvaluateAs(true, FakeBinding{}))
		})

		It("true or false", func() {
			expr := LogOrExpr{
				BooleanExpr{true},
				BooleanExpr{false},
			}

			Expect(expr).To(EvaluateAs(true, FakeBinding{}))
		})

		It("false or true", func() {
			expr := LogOrExpr{
				BooleanExpr{false},
				BooleanExpr{true},
			}

			Expect(expr).To(EvaluateAs(true, FakeBinding{}))
		})

		It("false or false", func() {
			expr := LogOrExpr{
				BooleanExpr{false},
				BooleanExpr{false},
			}

			Expect(expr).To(EvaluateAs(false, FakeBinding{}))
		})
	})

	Context("when both sides fail", func() {
		It("fails", func() {
			expr := LogOrExpr{
				FailingExpr{},
				FailingExpr{},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the left-hand side fails", func() {
		It("fails", func() {
			expr := LogOrExpr{
				FailingExpr{},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side fails", func() {
		It("fails", func() {
			expr := LogOrExpr{
				IntegerExpr{1},
				FailingExpr{},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the left-hand side is nil", func() {
		It("returns the right-hand side", func() {
			expr := LogOrExpr{
				NilExpr{},
				BooleanExpr{true},
			}

			Expect(expr).To(EvaluateAs(true, FakeBinding{}))
		})
	})

	Context("when the right side is nil", func() {
		It("returns the left-hand side", func() {
			expr := LogOrExpr{
				BooleanExpr{true},
				NilExpr{},
			}

			Expect(expr).To(EvaluateAs(true, FakeBinding{}))
		})
	})

	Context("when both sides are integers", func() {
		It("evaluates bit-wise or", func() {
			expr := LogOrExpr{
				IntegerExpr{5},
				IntegerExpr{6},
			}

			Expect(expr).To(EvaluateAs(7, FakeBinding{}))
		})
	})
})

package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("logical and", func() {
	Context("on boolean", func() {
		It("true and true", func() {
			expr := LogAndExpr{
				BooleanExpr{true},
				BooleanExpr{true},
			}

			Expect(expr).To(EvaluateAs(true, FakeBinding{}))
		})

		It("true and false", func() {
			expr := LogAndExpr{
				BooleanExpr{true},
				BooleanExpr{false},
			}

			Expect(expr).To(EvaluateAs(false, FakeBinding{}))
		})

		It("false name true", func() {
			expr := LogAndExpr{
				BooleanExpr{false},
				BooleanExpr{true},
			}

			Expect(expr).To(EvaluateAs(false, FakeBinding{}))
		})

		It("false and false", func() {
			expr := LogAndExpr{
				BooleanExpr{false},
				BooleanExpr{false},
			}

			Expect(expr).To(EvaluateAs(false, FakeBinding{}))
		})
	})

	Context("when both sides fail", func() {
		It("fails", func() {
			expr := LogAndExpr{
				FailingExpr{},
				FailingExpr{},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the left-hand side fails", func() {
		It("fails", func() {
			expr := LogAndExpr{
				FailingExpr{},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side fails", func() {
		It("fails", func() {
			expr := LogAndExpr{
				IntegerExpr{1},
				FailingExpr{},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the left-hand side is nil", func() {
		It("returns false", func() {
			expr := LogAndExpr{
				NilExpr{},
				BooleanExpr{true},
			}

			Expect(expr).To(EvaluateAs(false, FakeBinding{}))
		})
	})

	Context("when the right side is nil", func() {
		It("returns false", func() {
			expr := LogAndExpr{
				BooleanExpr{true},
				NilExpr{},
			}

			Expect(expr).To(EvaluateAs(false, FakeBinding{}))
		})
	})

	Context("when both sides are integers", func() {
		It("evaluates bit-wise and", func() {
			expr := LogAndExpr{
				IntegerExpr{5},
				IntegerExpr{6},
			}

			Expect(expr).To(EvaluateAs(4, FakeBinding{}))
		})
	})
})

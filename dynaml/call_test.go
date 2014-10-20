package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Builtin call", func() {
	It("handles bad names correctly", func() {
		Expect(CallExpr{Name: "function_does_not_exist", Arguments: []Expression{}}).To(FailToEvaluate(FakeBinding{}))
	})

	It("evaluates simple integer expressions correctly", func() {
		Expect(CallExpr{Name: "phase0_int_id", Arguments: []Expression{IntegerExpr{25}}}).To(EvaluateAs(25, FakeBinding{}))
	})

	It("evaluates simple string expressions correctly", func() {
		Expect(CallExpr{Name: "phase0_string_id", Arguments: []Expression{StringExpr{"foo"}}}).To(EvaluateAs("foo", FakeBinding{}))
	})

	It("catches type mismatches", func() {
		Expect(CallExpr{Name: "phase0_int_id", Arguments: []Expression{StringExpr{"foo"}}}).To(FailToEvaluate(FakeBinding{}))
	})

	It("does not evaluate functions require a phase to elapse", func() {
		expr := CallExpr{Name: "phase1_int_id", Arguments: []Expression{IntegerExpr{25}}}
		Expect(expr).To(DelayEvaluate(FakeBinding{}))
	})

	It("does evaluate functions after their phase has elapsed", func() {
		expr := CallExpr{Name: "phase1_int_id", Arguments: []Expression{IntegerExpr{25}}}
		binding := FakeBinding{}
		binding.ProvidedPhases.Add("phase1")
		Expect(expr).To(EvaluateAs(25, binding))
	})
})

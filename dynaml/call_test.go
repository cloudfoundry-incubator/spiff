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
})

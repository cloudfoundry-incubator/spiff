package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("calls", func() {
	It("it handles bad names correctly", func() {
		Expect(CallExpr{Name: "function_does_not_exist", Arguments: []Expression{}}).To(FailToEvaluate(FakeBinding{}))
	})

	It("it calls simple expressions correctly", func() {
		Expect(CallExpr{Name: "phase0_int_id", Arguments: []Expression{IntegerExpr{25}}}).To(EvaluateAs(25, FakeBinding{}))
	})
})

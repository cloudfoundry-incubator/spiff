package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("booleans", func() {
	It("evaluates to a bool", func() {
		Expect(BooleanExpr{true}).To(EvaluateAs(true, FakeBinding{}))
		Expect(BooleanExpr{false}).To(EvaluateAs(false, FakeBinding{}))
	})
})

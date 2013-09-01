package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("booleans", func() {
	d.It("evaluates to a bool", func() {
		Expect(BooleanExpr{true}).To(EvaluateAs(true, FakeContext{}))
		Expect(BooleanExpr{false}).To(EvaluateAs(false, FakeContext{}))
	})
})

package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("booleans", func() {
	d.It("evaluates to a bool", func() {
		Expect(BooleanExpr{true}.Evaluate(FakeContext{})).To(Equal(true))
		Expect(BooleanExpr{false}.Evaluate(FakeContext{})).To(Equal(false))
	})
})

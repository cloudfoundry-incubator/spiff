package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("integers", func() {
	d.It("evaluates to an int", func() {
		Expect(IntegerExpr{42}.Evaluate(FakeContext{})).To(Equal(42))
	})
})

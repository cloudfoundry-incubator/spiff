package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("integers", func() {
	It("evaluates to an int", func() {
		Expect(IntegerExpr{42}).To(EvaluateAs(int64(42), FakeBinding{}))
	})
})

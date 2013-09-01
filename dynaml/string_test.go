package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("integers", func() {
	d.It("evaluates to an int", func() {
		Expect(StringExpr{"foo"}).To(EvaluateAs("foo", FakeContext{}))
	})
})

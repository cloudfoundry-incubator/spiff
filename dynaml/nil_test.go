package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("nil", func() {
	d.It("evaluates to nil", func() {
		Expect(NilExpr{}).To(EvaluateAs(nil, FakeContext{}))
	})
})

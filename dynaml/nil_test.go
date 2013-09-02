package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("nil", func() {
	It("evaluates to nil", func() {
		Expect(NilExpr{}).To(EvaluateAs(nil, FakeBinding{}))
	})
})

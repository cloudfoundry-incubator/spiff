package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("logical and", func() {
	Context("on boolean", func() {
		It("not true = false", func() {
			expr := NotExpr{
				BooleanExpr{true},
			}

			Expect(expr).To(EvaluateAs(false, FakeBinding{}))
		})

		It("not false = true", func() {
			expr := NotExpr{
				BooleanExpr{false},
			}

			Expect(expr).To(EvaluateAs(true, FakeBinding{}))
		})
	})

	Context("on integer", func() {
		It("!=0 returns false", func() {
			expr := NotExpr{
				IntegerExpr{6},
			}

			Expect(expr).To(EvaluateAs(false, FakeBinding{}))
		})
	})
})

package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("subtractioj", func() {
	d.It("subtracts both numbers", func() {
		expr := ConcatenationExpr{
			StringExpr{"one"},
			StringExpr{"two"},
		}

		Expect(expr.Evaluate(FakeContext{})).To(Equal("onetwo"))
	})

	d.Context("when the left-hand side is not a string", func() {
		d.It("returns nil", func() {
			expr := ConcatenationExpr{
				StringExpr{"one"},
				IntegerExpr{42},
			}

			Expect(expr.Evaluate(FakeContext{})).To(BeNil())
		})
	})

	d.Context("when the right-hand side is not a string", func() {
		d.It("returns nil", func() {
			expr := ConcatenationExpr{
				IntegerExpr{42},
				StringExpr{"two"},
			}

			Expect(expr.Evaluate(FakeContext{})).To(BeNil())
		})
	})
})

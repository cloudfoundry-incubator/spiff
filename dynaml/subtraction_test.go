package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("subtraction", func() {
	d.It("subtracts both numbers", func() {
		expr := SubtractionExpr{
			IntegerExpr{7},
			IntegerExpr{3},
		}

		Expect(expr.Evaluate(FakeContext{})).To(Equal(4))
	})

	d.Context("when the left-hand side is not an integer", func() {
		d.It("returns nil", func() {
			expr := SubtractionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr.Evaluate(FakeContext{})).To(BeNil())
		})
	})

	d.Context("when the right-hand side is not an integer", func() {
		d.It("returns nil", func() {
			expr := SubtractionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr.Evaluate(FakeContext{})).To(BeNil())
		})
	})
})

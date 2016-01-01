package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("subtraction", func() {
	It("subtracts both numbers", func() {
		expr := SubtractionExpr{
			IntegerExpr{7},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs(4, FakeBinding{}))
	})

	Context("when the left-hand side is not an integer", func() {
		It("fails", func() {
			expr := SubtractionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is not an integer", func() {
		It("fails", func() {
			expr := SubtractionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
	
	Context("when the left-hand side is an IP address", func() {
		It("subtracts from the IP address without carry", func() {
			expr := SubtractionExpr{
				StringExpr{"10.10.10.10"},
				IntegerExpr{1},
			}

			Expect(expr).To(EvaluateAs("10.10.10.9",FakeBinding{}))
		})
		
		It("adds to the IP address with single byte carry", func() {
			expr := SubtractionExpr{
				StringExpr{"10.10.10.10"},
				IntegerExpr{257},
			}

			Expect(expr).To(EvaluateAs("10.10.9.9",FakeBinding{}))
		})
	})
})

package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("addition", func() {
	It("adds both numbers", func() {
		expr := AdditionExpr{
			IntegerExpr{2},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs(5, FakeBinding{}))
	})

	Context("when the left-hand side is not an integer", func() {
		It("fails", func() {
			expr := AdditionExpr{
				StringExpr{"lol"},
				IntegerExpr{2},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})

	Context("when the right-hand side is not an integer", func() {
		It("fails", func() {
			expr := AdditionExpr{
				IntegerExpr{2},
				StringExpr{"lol"},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
	
	Context("when the left-hand side is an IP address", func() {
		It("adds to the IP address without carry", func() {
			expr := AdditionExpr{
				StringExpr{"10.10.10.10"},
				IntegerExpr{1},
			}

			Expect(expr).To(EvaluateAs("10.10.10.11",FakeBinding{}))
		})
		
		It("adds to the IP address with single byte carry", func() {
			expr := AdditionExpr{
				StringExpr{"10.10.10.10"},
				IntegerExpr{257},
			}

			Expect(expr).To(EvaluateAs("10.10.11.11",FakeBinding{}))
		})
		
		It("adds to the IP address with simgle two byte carry", func() {
			expr := AdditionExpr{
				StringExpr{"10.10.10.10"},
				IntegerExpr{246+255*256},
			}

			Expect(expr).To(EvaluateAs("10.11.10.0",FakeBinding{}))
		})
		
		It("adds negative offset to the IP address without carry", func() {
			expr := AdditionExpr{
				StringExpr{"10.10.10.10"},
				IntegerExpr{-1},
			}

			Expect(expr).To(EvaluateAs("10.10.10.9",FakeBinding{}))
		})
		
		It("adds negative offset to the IP address with byte carry", func() {
			expr := AdditionExpr{
				StringExpr{"10.10.10.10"},
				IntegerExpr{-257},
			}

			Expect(expr).To(EvaluateAs("10.10.9.9",FakeBinding{}))
		})
		
		It("adds negative offset to the IP address with two byte carry", func() {
			expr := AdditionExpr{
				StringExpr{"10.10.10.10"},
				IntegerExpr{-11-65536},
			}

			Expect(expr).To(EvaluateAs("10.9.9.255",FakeBinding{}))
		})
	})
})

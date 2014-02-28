package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("concatenation", func() {
	Context("when the left-hand side is a string", func() {
		Context("and the right-hand side is a string", func() {
			It("concatenates both strings", func() {
				expr := ConcatenationExpr{
					StringExpr{"one"},
					StringExpr{"two"},
				}

				Expect(expr).To(EvaluateAs("onetwo", FakeBinding{}))
			})
		})

		Context("and the right-hand side is NOT a string", func() {
			It("fails", func() {
				expr := ConcatenationExpr{
					StringExpr{"two"},
					IntegerExpr{42},
				}

				Expect(expr).To(FailToEvaluate(FakeBinding{}))
			})
		})
	})

	Context("when the left-hand side is a list", func() {
		Context("and the right-hand side is a list", func() {
			It("concatenates both lists", func() {
				expr := ConcatenationExpr{
					ListExpr{[]Expression{StringExpr{"one"}}},
					ListExpr{[]Expression{StringExpr{"two"}}},
				}

				Expect(expr).To(EvaluateAs([]yaml.Node{node("one"), node("two")}, FakeBinding{}))
			})
		})

		Context("and the right-hand side is NOT a list", func() {
			It("fails", func() {
				expr := ConcatenationExpr{
					IntegerExpr{42},
					ListExpr{[]Expression{StringExpr{"two"}}},
				}

				Expect(expr).To(FailToEvaluate(FakeBinding{}))
			})
		})
	})
})

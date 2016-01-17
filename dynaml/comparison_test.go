package dynaml

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func compareIt(op string, result []bool) {
	for first := 5; first <= 7; first++ {
		expected := result[first-5]
		arg := int64(first)
		It(fmt.Sprintf("evaluates %d%s6", first, op), func() {
			expr := ComparisonExpr{
				IntegerExpr{arg},
				op,
				IntegerExpr{6},
			}

			Expect(expr).To(EvaluateAs(expected, FakeBinding{}))
		})
	}
}

var _ = Describe("comparison operators", func() {
	Context("<=", func() {
		compareIt("<=", []bool{true, true, false})
		compareIt("<", []bool{true, false, false})
		compareIt(">=", []bool{false, true, true})
		compareIt(">", []bool{false, false, true})
		compareIt("==", []bool{false, true, false})
		compareIt("!=", []bool{true, false, true})
	})

	Context("when one side fails", func() {
		It("fails for left side failing", func() {
			expr := ComparisonExpr{
				FailingExpr{},
				"<=",
				IntegerExpr{6},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})

		It("fails for right side failing", func() {
			expr := ComparisonExpr{
				IntegerExpr{6},
				"<=",
				FailingExpr{},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
})

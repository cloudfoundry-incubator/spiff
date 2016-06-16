package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var _ = Describe("integer range", func() {
	It("evaluates an increasing range", func() {
		expr := RangeExpr{
			IntegerExpr{1},
			IntegerExpr{3},
		}

		Expect(expr).To(EvaluateAs([]yaml.Node{node(1, nil), node(2, nil), node(3, nil)}, FakeBinding{}))
	})

	It("evaluates a decreasing range", func() {
		expr := RangeExpr{
			IntegerExpr{1},
			IntegerExpr{-1},
		}

		Expect(expr).To(EvaluateAs([]yaml.Node{node(1, nil), node(0, nil), node(-1, nil)}, FakeBinding{}))
	})

	It("evaluates a single element range", func() {
		expr := RangeExpr{
			IntegerExpr{1},
			IntegerExpr{1},
		}

		Expect(expr).To(EvaluateAs([]yaml.Node{node(1, nil)}, FakeBinding{}))
	})

	It("evaluates to failure", func() {
		expr := RangeExpr{
			StringExpr{"foo"},
			IntegerExpr{1},
		}

		Expect(expr).To(FailToEvaluate(FakeBinding{}))
	})
})

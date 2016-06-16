package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var _ = Describe("lists", func() {
	It("evaluates to an array of nodes", func() {
		expr := ListExpr{
			[]Expression{
				IntegerExpr{1},
				StringExpr{"two"},
			},
		}

		Expect(expr).To(EvaluateAs([]yaml.Node{node(1, nil), node("two", nil)}, FakeBinding{}))
	})

	Context("when empty", func() {
		It("evaluates to an empty array", func() {
			Expect(ListExpr{}).To(EvaluateAs([]yaml.Node{}, FakeBinding{}))
		})
	})

	Context("when an entry does not resolve", func() {
		It("fails", func() {
			expr := ListExpr{
				[]Expression{
					ReferenceExpr{[]string{"foo"}},
				},
			}

			Expect(expr).To(FailToEvaluate(FakeBinding{}))
		})
	})
})

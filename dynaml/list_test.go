package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/yaml"
)

var _ = Describe("lists", func() {
	It("evaluates to an array of nodes", func() {
		expr := ListExpr{
			[]Expression{
				IntegerExpr{1},
				StringExpr{"two"},
			},
		}

		Expect(expr).To(EvaluateAs([]yaml.Node{1, "two"}, FakeBinding{}))
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

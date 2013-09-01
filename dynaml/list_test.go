package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/yaml"
)

var _ = d.Describe("lists", func() {
	d.It("evaluates to an array of nodes", func() {
		expr := ListExpr{
			[]Expression{
				IntegerExpr{1},
				StringExpr{"two"},
			},
		}

		Expect(expr).To(EvaluateAs([]yaml.Node{1, "two"}, FakeContext{}))
	})

	d.Context("when an entry does not resolve", func() {
		d.It("fails", func() {
			expr := ListExpr{
				[]Expression{
					ReferenceExpr{[]string{"foo"}},
				},
			}

			Expect(expr).To(FailToEvaluate(FakeContext{}))
		})
	})
})

package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/yaml"
)

var _ = Describe("references", func() {
	Context("when the reference is found", func() {
		It("evaluates to the referenced node", func() {
			expr := ReferenceExpr{[]string{"foo", "bar"}}

			binding := FakeBinding{
				FoundReferences: map[string]yaml.Node{
					"foo.bar": 42,
				},
			}

			Expect(expr).To(EvaluateAs(42, binding))
		})

		Context("and it refers to another expression", func() {
			It("fails", func() {
				referencedNode := IntegerExpr{42}

				expr := ReferenceExpr{[]string{"foo", "bar"}}

				binding := FakeBinding{
					FoundReferences: map[string]yaml.Node{
						"foo.bar": referencedNode,
					},
				}

				Expect(expr).To(FailToEvaluate(binding))
			})
		})
	})

	Context("when the reference is NOT found", func() {
		It("fails", func() {
			referencedNode := IntegerExpr{42}

			expr := ReferenceExpr{[]string{"foo", "bar", "baz"}}

			binding := FakeBinding{
				FoundReferences: map[string]yaml.Node{
					"foo.bar": referencedNode,
				},
			}

			Expect(expr).To(FailToEvaluate(binding))
		})
	})
})

package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/yaml"
)

var _ = d.Describe("references", func() {
	d.Context("when the reference is found", func() {
		d.It("evaluates to the referenced node", func() {
			expr := ReferenceExpr{[]string{"foo", "bar"}}

			context := FakeContext{
				FoundReferences: map[string]yaml.Node{
					"foo.bar": 42,
				},
			}

			Expect(expr).To(EvaluateAs(42, context))
		})

		d.Context("and it refers to another expression", func() {
			d.It("fails", func() {
				referencedNode := IntegerExpr{42}

				expr := ReferenceExpr{[]string{"foo", "bar"}}

				context := FakeContext{
					FoundReferences: map[string]yaml.Node{
						"foo.bar": referencedNode,
					},
				}

				Expect(expr).To(FailToEvaluate(context))
			})
		})
	})

	d.Context("when the reference is NOT found", func() {
		d.It("fails", func() {
			referencedNode := IntegerExpr{42}

			expr := ReferenceExpr{[]string{"foo", "bar", "baz"}}

			context := FakeContext{
				FoundReferences: map[string]yaml.Node{
					"foo.bar": referencedNode,
				},
			}

			Expect(expr).To(FailToEvaluate(context))
		})
	})
})

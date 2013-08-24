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

			Expect(expr.Evaluate(context)).To(Equal(42))
		})

		d.Context("and it refers to another expression", func() {
			d.It("returns nil", func() {
				referencedNode := IntegerExpr{42}

				expr := ReferenceExpr{[]string{"foo", "bar"}}

				context := FakeContext{
					FoundReferences: map[string]yaml.Node{
						"foo.bar": referencedNode,
					},
				}

				Expect(expr.Evaluate(context)).To(BeNil())
			})
		})
	})

	d.Context("when the reference is NOT found", func() {
		d.It("evaluates to nil", func() {
			referencedNode := IntegerExpr{42}

			expr := ReferenceExpr{[]string{"foo", "bar", "baz"}}

			context := FakeContext{
				FoundReferences: map[string]yaml.Node{
					"foo.bar": referencedNode,
				},
			}

			Expect(expr.Evaluate(context)).To(BeNil())
		})
	})
})

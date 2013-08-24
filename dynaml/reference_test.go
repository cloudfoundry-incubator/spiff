package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("references", func() {
	d.Context("when the reference is found", func() {
		d.It("evaluates to the referenced node", func() {
			referencedNode := IntegerExpr{42}

			expr := ReferenceExpr{[]string{"foo", "bar"}}

			context := FakeContext{
				FoundReferences: map[string]Node{
					"foo.bar": referencedNode,
				},
			}

			Expect(expr.Evaluate(context)).To(Equal(referencedNode))
		})
	})

	d.Context("when the reference is NOT found", func() {
		d.It("evaluates to nil", func() {
			referencedNode := IntegerExpr{42}

			expr := ReferenceExpr{[]string{"foo", "bar", "baz"}}

			context := FakeContext{
				FoundReferences: map[string]Node{
					"foo.bar": referencedNode,
				},
			}

			Expect(expr.Evaluate(context)).To(BeNil())
		})
	})
})

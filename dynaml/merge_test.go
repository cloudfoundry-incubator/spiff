package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/yaml"
)

var _ = d.Describe("merges", func() {
	d.Context("when the equivalent node is found", func() {
		d.It("evaluates to the merged node", func() {
			referencedNode := IntegerExpr{42}

			expr := MergeExpr{[]string{"foo", "bar"}}

			context := FakeContext{
				FoundInStubs: map[string]yaml.Node{
					"foo.bar": referencedNode,
				},
			}

			Expect(expr).To(EvaluateAs(referencedNode, context))
		})
	})

	d.Context("when the equivalent node is NOT found", func() {
		d.It("fails", func() {
			referencedNode := IntegerExpr{42}

			expr := MergeExpr{[]string{"foo", "bar", "baz"}}

			context := FakeContext{
				FoundInStubs: map[string]yaml.Node{
					"foo.bar": referencedNode,
				},
			}

			Expect(expr).To(FailToEvaluate(context))
		})
	})
})

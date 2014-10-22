package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/shutej/spiff/yaml"
)

var _ = Describe("merges", func() {
	Context("when the equivalent node is found", func() {
		It("evaluates to the merged node", func() {
			referencedNode := IntegerExpr{42}

			expr := MergeExpr{[]string{"foo", "bar"}}

			binding := FakeBinding{
				FoundInStubs: map[string]yaml.Node{
					"foo.bar": node(referencedNode),
				},
			}

			Expect(expr).To(EvaluateAs(referencedNode, binding))
		})
	})

	Context("when the equivalent node is NOT found", func() {
		It("fails", func() {
			referencedNode := IntegerExpr{42}

			expr := MergeExpr{[]string{"foo", "bar", "baz"}}

			binding := FakeBinding{
				FoundInStubs: map[string]yaml.Node{
					"foo.bar": node(referencedNode),
				},
			}

			Expect(expr).To(FailToEvaluate(binding))
		})
	})
})

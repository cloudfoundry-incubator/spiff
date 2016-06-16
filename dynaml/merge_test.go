package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var _ = Describe("merges", func() {
	Context("when the equivalent node is found", func() {
		It("evaluates to the merged node", func() {
			referencedNode := IntegerExpr{42}

			expr := MergeExpr{
				Path:     []string{"foo", "bar"},
				Redirect: false,
				Replace:  false,
				Required: false,
				KeyName:  "",
			}

			binding := FakeBinding{
				FoundInStubs: map[string]yaml.Node{
					"foo.bar": node(referencedNode, nil),
				},
			}

			Expect(expr).To(EvaluateAs(referencedNode, binding))
		})
	})

	Context("when the equivalent node is NOT found", func() {
		It("fails", func() {
			referencedNode := IntegerExpr{42}

			expr := MergeExpr{
				Path:     []string{"foo", "bar", "baz"},
				Redirect: false,
				Replace:  false,
				Required: false,
				KeyName:  "",
			}

			binding := FakeBinding{
				FoundInStubs: map[string]yaml.Node{
					"foo.bar": node(referencedNode, nil),
				},
			}

			Expect(expr).To(FailToEvaluate(binding))
		})
	})
})

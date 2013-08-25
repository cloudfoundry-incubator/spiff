package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/yaml"
)

var _ = d.Describe("calls", func() {
	d.Describe("static_ips(ips...)", func() {
		expr := CallExpr{
			Name: "static_ips",
			Arguments: []Expression{
				IntegerExpr{0},
				IntegerExpr{4},
			},
		}

		d.It("returns a set of ips from the given network", func() {
			static := parseYAML(`
- 10.10.16.10 - 10.10.16.254
`)

			context := FakeContext{
				FoundReferences: map[string]yaml.Node{
					"name": "cf1",
				},
				FoundFromRoot: map[string]yaml.Node{
					"networks.cf1.subnets.[0].static": static,
				},
			}

			Expect(expr.Evaluate(context)).To(Equal([]yaml.Node{
				"10.10.16.10",
				"10.10.16.14",
			}))
		})

		d.It("limits the IPs to the number of instances", func() {
			static := parseYAML(`
- 10.10.16.10 - 10.10.16.254
`)

			context := FakeContext{
				FoundReferences: map[string]yaml.Node{
					"name":      "cf1",
					"instances": 1,
				},
				FoundFromRoot: map[string]yaml.Node{
					"networks.cf1.subnets.[0].static": static,
				},
			}

			Expect(expr.Evaluate(context)).To(Equal([]yaml.Node{
				"10.10.16.10",
			}))
		})

		d.Context("when the instance count is dynamic", func() {
			d.It("returns nil", func() {
				static := parseYAML(`
- 10.10.16.10 - 10.10.16.254
`)

				context := FakeContext{
					FoundReferences: map[string]yaml.Node{
						"name":      "cf1",
						"instances": MergeExpr{},
					},
					FoundFromRoot: map[string]yaml.Node{
						"networks.cf1.subnets.[0].static": static,
					},
				}

				Expect(expr.Evaluate(context)).To(BeNil())
			})
		})
	})
})

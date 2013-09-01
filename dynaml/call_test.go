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
					"name":      "cf1",
					"instances": 2,
				},
				FoundFromRoot: map[string]yaml.Node{
					"networks.cf1.subnets.[0].static": static,
				},
			}

			Expect(expr).To(
				EvaluateAs(
					[]yaml.Node{"10.10.16.10", "10.10.16.14"},
					context,
				),
			)
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

			Expect(expr).To(
				EvaluateAs(
					[]yaml.Node{"10.10.16.10"},
					context,
				),
			)
		})

		d.Context("when the instance count is dynamic", func() {
			d.It("fails", func() {
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

				Expect(expr).To(FailToEvaluate(context))
			})
		})

		d.Context("when there are not enough IPs for the number of instances", func() {
			d.It("fails", func() {
				static := parseYAML(`
- 10.10.16.10 - 10.10.16.32
`)

				context := FakeContext{
					FoundReferences: map[string]yaml.Node{
						"name":      "cf1",
						"instances": 42,
					},
					FoundFromRoot: map[string]yaml.Node{
						"networks.cf1.subnets.[0].static": static,
					},
				}

				Expect(expr).To(FailToEvaluate(context))
			})
		})
	})
})

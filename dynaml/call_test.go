package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/yaml"
)

var _ = Describe("calls", func() {
	Describe("static_ips(ips...)", func() {
		expr := CallExpr{
			Name: "static_ips",
			Arguments: []Expression{
				IntegerExpr{0},
				IntegerExpr{4},
			},
		}

		It("returns a set of ips from the given network", func() {
			static := parseYAML(`
- 10.10.16.10 - 10.10.16.254
`)

			binding := FakeBinding{
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
					binding,
				),
			)
		})

		It("limits the IPs to the number of instances", func() {
			static := parseYAML(`
- 10.10.16.10 - 10.10.16.254
`)

			binding := FakeBinding{
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
					binding,
				),
			)
		})

		Context("when the instance count is dynamic", func() {
			It("fails", func() {
				static := parseYAML(`
- 10.10.16.10 - 10.10.16.254
`)

				binding := FakeBinding{
					FoundReferences: map[string]yaml.Node{
						"name":      "cf1",
						"instances": MergeExpr{},
					},
					FoundFromRoot: map[string]yaml.Node{
						"networks.cf1.subnets.[0].static": static,
					},
				}

				Expect(expr).To(FailToEvaluate(binding))
			})
		})

		Context("when there are not enough IPs for the number of instances", func() {
			It("fails", func() {
				static := parseYAML(`
- 10.10.16.10 - 10.10.16.32
`)

				binding := FakeBinding{
					FoundReferences: map[string]yaml.Node{
						"name":      "cf1",
						"instances": 42,
					},
					FoundFromRoot: map[string]yaml.Node{
						"networks.cf1.subnets.[0].static": static,
					},
				}

				Expect(expr).To(FailToEvaluate(binding))
			})
		})

		Context("when there are singular static IPs listed", func() {
			It("includes them in the pool", func() {
				static := parseYAML(`
- 10.10.16.10 - 10.10.16.32
- 10.10.16.33
- 10.10.16.34
`)

				expr := CallExpr{
					Name: "static_ips",
					Arguments: []Expression{
						IntegerExpr{0},
						IntegerExpr{4},
						IntegerExpr{23},
					},
				}

				binding := FakeBinding{
					FoundReferences: map[string]yaml.Node{
						"name":      "cf1",
						"instances": 3,
					},
					FoundFromRoot: map[string]yaml.Node{
						"networks.cf1.subnets.[0].static": static,
					},
				}

				Expect(expr).To(
					EvaluateAs(
						[]yaml.Node{"10.10.16.10", "10.10.16.14", "10.10.16.33"},
						binding,
					),
				)
			})
		})
	})
})

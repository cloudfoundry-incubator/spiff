package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/yaml"
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

		It("returns a set of ips from the given network's subnets", func() {
			subnets := parseYAML(`
- static:
    - 10.10.16.10
- static:
    - 10.10.16.11 - 10.10.16.254
`)

			binding := FakeBinding{
				FoundReferences: map[string]yaml.Node{
					"name":      node("cf1"),
					"instances": node(2),
				},
				FoundFromRoot: map[string]yaml.Node{
					"networks.cf1.subnets": subnets,
				},
			}

			Expect(expr).To(
				EvaluateAs(
					[]yaml.Node{node("10.10.16.10"), node("10.10.16.14")},
					binding,
				),
			)
		})

		It("limits the IPs to the number of instances", func() {
			subnets := parseYAML(`
- static:
    - 10.10.16.10 - 10.10.16.254
`)

			binding := FakeBinding{
				FoundReferences: map[string]yaml.Node{
					"name":      node("cf1"),
					"instances": node(1),
				},
				FoundFromRoot: map[string]yaml.Node{
					"networks.cf1.subnets": subnets,
				},
			}

			Expect(expr).To(
				EvaluateAs(
					[]yaml.Node{node("10.10.16.10")},
					binding,
				),
			)
		})

		Context("when the instance count is dynamic", func() {
			It("fails", func() {
				subnets := parseYAML(`
- static:
    - 10.10.16.10 - 10.10.16.254
`)

				binding := FakeBinding{
					FoundReferences: map[string]yaml.Node{
						"name":      node("cf1"),
						"instances": node(MergeExpr{}),
					},
					FoundFromRoot: map[string]yaml.Node{
						"networks.cf1.subnets": subnets,
					},
				}

				Expect(expr).To(FailToEvaluate(binding))
			})
		})

		Context("when there are not enough IPs for the number of instances", func() {
			It("fails", func() {
				subnets := parseYAML(`
- static:
    - 10.10.16.10 - 10.10.16.32
`)

				binding := FakeBinding{
					FoundReferences: map[string]yaml.Node{
						"name":      node("cf1"),
						"instances": node(42),
					},
					FoundFromRoot: map[string]yaml.Node{
						"networks.cf1.subnets": subnets,
					},
				}

				Expect(expr).To(FailToEvaluate(binding))
			})
		})

		Context("when there are singular static IPs listed", func() {
			It("includes them in the pool", func() {
				subnets := parseYAML(`
- static:
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
						"name":      node("cf1"),
						"instances": node(3),
					},
					FoundFromRoot: map[string]yaml.Node{
						"networks.cf1.subnets": subnets,
					},
				}

				Expect(expr).To(
					EvaluateAs(
						[]yaml.Node{node("10.10.16.10"), node("10.10.16.14"), node("10.10.16.33")},
						binding,
					),
				)
			})
		})

		Context("when there is IP range in descending order", func() {
			It("includes IPs in the pool in ascending order", func() {
				subnets := parseYAML(`
- static:
    - 10.10.16.10 - 10.10.16.1
`)

				expr := CallExpr{
					Name: "static_ips",
					Arguments: []Expression{
						IntegerExpr{0},
						IntegerExpr{1},
						IntegerExpr{2},
					},
				}

				binding := FakeBinding{
					FoundReferences: map[string]yaml.Node{
						"name":      node("cf1"),
						"instances": node(3),
					},
					FoundFromRoot: map[string]yaml.Node{
						"networks.cf1.subnets": subnets,
					},
				}

				Expect(expr).To(
					EvaluateAs(
						[]yaml.Node{node("10.10.16.1"), node("10.10.16.2"), node("10.10.16.3")},
						binding,
					),
				)
			})
		})
	})
})

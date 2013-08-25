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
			network := parseYAML(`
type: manual
subnets:
  - range: 10.10.16.0/20
    name: default_unused
    reserved:
      - 10.10.16.2 - 10.10.16.9
      - 10.10.16.255 - 10.10.16.255
    static:
      - 10.10.16.10 - 10.10.16.254
    gateway: 10.10.16.1
    dns:
      - 10.10.0.2`)

			context := FakeContext{
				FoundReferences: map[string]yaml.Node{
					"name": "cf1",
				},
				FoundFromRoot: map[string]yaml.Node{
					"networks.cf1": network,
				},
			}

			Expect(expr.Evaluate(context)).To(Equal([]yaml.Node{
				"10.10.16.10",
				"10.10.16.14",
			}))
		})
	})
})

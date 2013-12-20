package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/dynaml"
)

var _ = Describe("Flowing YAML", func() {
	Context("when there are no dynaml nodes", func() {
		It("is a no-op", func() {
			source := parseYAML(`
---
foo: bar
`)

			Expect(source).To(FlowAs(source))
		})
	})

	Context("when there are no dynaml nodes", func() {
		It("is a no-op", func() {
			source := parseYAML(`
---
foo: bar
`)

			Expect(source).To(FlowAs(source))
		})
	})

	Context("when a value is defined in the template and a stub", func() {
		It("overrides the value with the stubbed value", func() {
			source := parseYAML(`
---
a: ~
b: 1
c: foo
d: 2.5
fizz: buzz
`)

			stub := parseYAML(`
---
a: b
b: 2
c: bar
d: 3.14
`)

			result := parseYAML(`
---
a: b
b: 2
c: bar
d: 3.14
fizz: buzz
`)
			Expect(source).To(FlowAs(result, stub))
		})
	})

	Context("when some dynaml nodes cannot be resolved", func() {
		It("returns an error", func() {
			source := parseYAML(`
---
foo: (( auto ))
`)

			_, err := Flow(source)
			Expect(err).To(Equal(UnresolvedNodes{
				Nodes: []dynaml.Expression{
					dynaml.AutoExpr{
						Path: []string{"foo"},
					},
				},
			}))
		})
	})

	Context("when a reference is made to a yet-to-be-resolved node, in a || expression", func() {
		It("eventually resolves to the referenced node", func() {
			source := parseYAML(`
---
properties:
  template_only: (( merge ))
  something: (( template_only.foo || "wrong" ))
`)

			stub := parseYAML(`
---
properties:
  template_only:
    foo: right
`)

			resolved := parseYAML(`
---
properties:
  template_only:
    foo: right
  something: right
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Context("when a refence is made to an unresolveable node", func() {
		It("fails to flow", func() {
			source := parseYAML(`
---
properties:
  template_only: (( abc ))
  something: (( template_only.foo ))
`)

			_, err := Flow(source)
			Expect(err).To(HaveOccurred())
		})
	})

	Context("when a reference is made to an unresolveable node, in a || expression", func() {
		It("eventually resolves to the referenced node", func() {
			source := parseYAML(`
---
properties:
  template_only: (( merge ))
  something: (( template_only.foo || "right" ))
`)

			stub := parseYAML(`
---
properties:
  template_only:
`)

			resolved := parseYAML(`
---
properties:
  template_only:
  something: right
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("basic dynaml nodes", func() {
		It("evaluates the nodes", func() {
			source := parseYAML(`
---
foo:
  - (( "hello, world!" ))
  - (( 42 ))
  - (( true ))
  - (( nil ))
`)

			resolved := parseYAML(`
---
foo:
  - hello, world!
  - 42
  - true
  - null
`)

			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("reference dynaml nodes", func() {
		It("evaluates the node", func() {
			source := parseYAML(`
---
foo: (( bar ))
bar: 42
`)

			resolved := parseYAML(`
---
foo: 42
bar: 42
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("follows lexical scoping semantics", func() {
			source := parseYAML(`
---
foo:
  bar:
    baz: (( buzz.fizz ))
  buzz:
    fizz: right
buzz:
  fizz: wrong
`)

			resolved := parseYAML(`
---
foo:
  bar:
    baz: right
  buzz:
    fizz: right
buzz:
  fizz: wrong
`)

			Expect(source).To(FlowAs(resolved))
		})

		Context("when the referred node is dynamic", func() {
			It("evaluates with their environment", func() {
				source := parseYAML(`
---
foo:
  bar:
    baz: (( buzz.fizz ))
    quux: wrong
buzz:
  fizz: (( quux ))
  quux: right
`)

				resolved := parseYAML(`
---
foo:
  bar:
    baz: right
    quux: wrong
buzz:
  fizz: right
  quux: right
`)

				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("merging in from stubs", func() {
		It("evaluates the node", func() {
			source := parseYAML(`
---
foo: (( merge ))
bar: 42
`)

			stub := parseYAML(`
---
foo: merged!
`)

			resolved := parseYAML(`
---
foo: merged!
bar: 42
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("follows through maps in lists by name", func() {
			source := parseYAML(`
---
foo:
- name: x
  value: (( merge ))
`)

			stub := parseYAML(`
---
foo:
- name: y
  value: wrong
- name: x
  value: right
`)

			resolved := parseYAML(`
---
foo:
- name: x
  value: right
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		// this is a regression test, from when Environment.WithPath
		// used append() for adding the next step.
		//
		// using append() will overwrite previous steps, since it reuses the slice
		//
		// e.g. with inital path A:
		//    append(A, "a")
		//    append(A, "b")
		//
		// would result in all previous A/a paths becoming A/b
		It("can be arbitrarily nested", func() {
			source := parseYAML(`
---
properties:
  something:
    foo:
      key: (( merge ))
      val: (( merge ))
`)

			stub := parseYAML(`
---
properties:
  something:
    foo:
      key: a
      val: b
`)

			resolved := parseYAML(`
---
properties:
  something:
    foo:
      key: a
      val: b
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("automatic resource pool sizes", func() {
		It("evaluates the node", func() {
			source := parseYAML(`
---
resource_pools:
  some_pool:
    size: (( auto ))

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
- name: some_other_job
  resource_pool: some_pool
  instances: 3
- name: yet_another_job
  resource_pool: some_other_pool
  instances: 5
`)

			resolved := parseYAML(`
---
resource_pools:
  some_pool:
    size: 5

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
- name: some_other_job
  resource_pool: some_pool
  instances: 3
- name: yet_another_job
  resource_pool: some_other_pool
  instances: 5
`)

			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("static ip population", func() {
		It("evaluates the node", func() {
			source := parseYAML(`
---
networks:
  some_network:
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
          - 10.10.0.2

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
  networks:
  - name: some_network
    static_ips: (( static_ips(0, 4) ))
`)

			resolved := parseYAML(`
---
networks:
  some_network:
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
          - 10.10.0.2

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
  networks:
  - name: some_network
    static_ips:
    - 10.10.16.10
    - 10.10.16.14
`)

			Expect(source).To(FlowAs(resolved))
		})
	})
})

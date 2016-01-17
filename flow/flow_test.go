package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/dynaml"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var _ = Describe("Flowing YAML", func() {
	Context("delays resolution until merge succeeded", func() {
		It("handles combination of inline merge and field merge", func() {
			source := parseYAML(`
---
properties:
  <<: (( merge || nil ))
  bar: (( merge ))

foobar:
  - (( "foo." .properties.bar ))
`)
			stub := parseYAML(`
---
properties:
  bar: bar
`)

			resolved := parseYAML(`
---
properties:
  bar: bar
foobar: 
  - foo.bar
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("handles defaulted reference to merged/overridden fields", func() {
			source := parseYAML(`
---
foo:
  <<: (( merge || nil ))
  bar:
    <<: (( merge || nil ))
    alice: alice

props:
  bob: (( foo.bar.bob || "wrong" ))
  alice: (( foo.bar.alice || "wrong" ))
  main: (( foo.foo || "wrong" ))

`)
			stub := parseYAML(`
---
foo: 
  foo: added
  bar:
    alice: overwritten
    bob: added!

`)

			resolved := parseYAML(`
---
foo:
  bar:
    alice: overwritten
    bob: added!
  foo: added
props:
  alice: overwritten
  bob: added!
  main: added

`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("handles defaulted reference to merged/overridden redirected fields", func() {
			source := parseYAML(`
---
foo:
  <<: (( merge alt || nil ))
  bar:
    <<: (( merge || nil ))
    alice: alice

props:
  bob: (( foo.bar.bob || "wrong" ))
  alice: (( foo.bar.alice || "wrong" ))
  main: (( foo.foo || "wrong" ))

`)
			stub := parseYAML(`
---
foo:
  bar:
    alice: wrongly merged
alt:
  foo: added
  bar:
    alice: overwritten
    bob: added!

`)

			resolved := parseYAML(`
---
foo:
  bar:
    alice: overwritten
    bob: added!
  foo: added
props:
  alice: overwritten
  bob: added!
  main: added

`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("replaces a non-merge expression node before expanding", func() {
			source := parseYAML(`
---
alt:
  - wrong
properties: (( alt ))
`)
			stub := parseYAML(`
---
properties:
  - right
`)

			resolved := parseYAML(`
---
alt:
  - wrong
properties:
  - right
`)
			Expect(source).To(FlowAs(resolved, stub))
		})

		It("expands a preferred non-merge expression node before overriding", func() {
			source := parseYAML(`
---
alt:
  - right
properties: (( prefer alt ))
`)
			stub := parseYAML(`
---
properties:
  - wrong
`)

			resolved := parseYAML(`
---
alt:
  - right
properties:
  - right
`)
			Expect(source).To(FlowAs(resolved, stub))
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

		Context("in a list", func() {
			It("does not override the value", func() {
				source := parseYAML(`
---
- 1
- 2
`)

				stub := parseYAML(`
---
- 3
- 4
`)

				Expect(source).To(FlowAs(source, stub))
			})
		})
	})

	Context("when some dynaml nodes cannot be resolved", func() {
		It("returns an error", func() {
			source := parseYAML(`
---
foo: (( auto ))
`)

			_, err := Flow(source)
			Expect(err).To(Equal(dynaml.UnresolvedNodes{
				Nodes: []dynaml.UnresolvedNode{
					{
						Node: yaml.IssueNode(yaml.NewNode(
							dynaml.AutoExpr{Path: []string{"foo"}},
							"test",
						), "auto only allowed for size entry in resource pools"),
						Context: []string{"foo"},
						Path:    []string{"foo"},
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

		Context("when the reference starts with .", func() {
			It("starts from the root of the template", func() {
				source := parseYAML(`
---
foo:
  bar:
    baz: (( .bar.buzz ))
    buzz: 42
bar:
  buzz: 43
`)

				resolved := parseYAML(`
---
foo:
  bar:
    baz: 43
    buzz: 42
bar:
  buzz: 43
`)

				Expect(source).To(FlowAs(resolved))
			})
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

	Describe("merging fields", func() {
		It("merges locally referenced fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( bar ))
  other: other
bar:
  alice: alice
  bob: bob
`)

			resolved := parseYAML(`
---
foo:
  alice: alice
  bob: bob
  other: other
bar:
  alice: alice
  bob: bob
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("overwrites locally referenced fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( bar ))
  alice: overwritten
  other: other
bar:
  alice: alice
  bob: bob
`)

			resolved := parseYAML(`
---
foo:
  alice: overwritten
  bob: bob
  other: other
bar:
  alice: alice
  bob: bob
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("merges redirected stub fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge alt ))
bar: 42
`)

			stub := parseYAML(`
---
foo: 
  alice: not merged!
alt: 
  bob: merged!
`)

			resolved := parseYAML(`
---
foo: 
  bob: merged!
bar: 42
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("overwrites redirected stub fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge alt ))
  bar: 42
`)

			stub := parseYAML(`
---
foo: 
  alice: not merged!
alt: 
  bob: added!
  bar: overwritten
`)

			resolved := parseYAML(`
---
foo: 
  bob: added!
  bar: overwritten
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("resolves overwritten redirected stub fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge alt ))
  bar: 42
ref:
  bar: (( foo.bar ))
`)

			stub := parseYAML(`
---
foo: 
  alice: not merged!
alt: 
  bob: added!
  bar: overwritten
`)

			resolved := parseYAML(`
---
foo: 
  bob: added!
  bar: overwritten
ref:
  bar: overwritten
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("deep overwrites redirected stub fields", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge alt ))
  bar:
    alice: alice
    bob: bob
`)

			stub := parseYAML(`
---
foo: 
  alice: not merged!
alt: 
  bob: added!
  bar:
    alice: overwritten
`)

			resolved := parseYAML(`
---
foo: 
  bar:
    alice: overwritten
    bob: bob
  bob: added!
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("propagates redirection to subsequent merges", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge alt ))
  bar:
    <<: (( merge ))
    alice: alice
`)

			stub := parseYAML(`
---
foo: 
  alice: not merged!
alt: 
  bar:
    alice: overwritten
    bob: added!
`)

			resolved := parseYAML(`
---
foo: 
  bar:
    alice: overwritten
    bob: added!
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	// replace whole structure instead of deep override
	Describe("replacing nodes from stubs", func() {
		It("does nothing for no direct match", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge replace || nil ))
  bar: 42
`)

			resolved := parseYAML(`
---
foo: 
  bar: 42
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("copies the node", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge replace ))
  bar: 42
`)

			stub := parseYAML(`
---
foo: 
  blah: replaced!
`)

			resolved := parseYAML(`
---
foo: 
  blah: replaced!
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("does not follow through maps in lists by name", func() {
			source := parseYAML(`
---
foo:
- <<: (( merge replace ))
- name: x
  value: v
`)

			stub := parseYAML(`
---
foo:
- name: y
  value: right
- name: z
  value: right
`)

			resolved := parseYAML(`
---
foo:
- name: y
  value: right
- name: z
  value: right
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("doesn't hamper field value merge", func() {
			source := parseYAML(`
---
foo:
  bar: (( merge replace ))
`)

			stub := parseYAML(`
---
foo:
  bar:
    value: right
`)

			resolved := parseYAML(`
---
foo:
  bar:
    value: right
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("doesn't hamper list value merge", func() {
			source := parseYAML(`
---
foo:
  bar: (( merge replace ))
`)

			stub := parseYAML(`
---
foo:
  bar:
    - alice
    - bob
`)

			resolved := parseYAML(`
---
foo:
  bar:
    - alice
    - bob
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("replacing map with redirection", func() {
		It("merges with redirected map, but not with original path", func() {
			source := parseYAML(`
---
foo: 
  <<: (( merge replace bar ))
  bar:
    alice: alice
    bob: bob
`)

			stub := parseYAML(`
---
foo:
  alice: not merged
bar:
  alice: merged
  bob: merged
`)

			resolved := parseYAML(`
---
foo:
  alice: merged
  bob: merged
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("replacing list with redirection", func() {
		It("merges with redirected map, but not with original path", func() {
			source := parseYAML(`
---
foo: 
  - <<: (( merge replace bar ))
  - bar:
      alice: alice
      bob: bob
`)

			stub := parseYAML(`
---
foo:
  - not
  - merged
bar:
  - alice: merged
  - bob: merged
`)

			resolved := parseYAML(`
---
foo:
  - alice: merged
  - bob: merged
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("resolves references to merges with redirected map", func() {
			source := parseYAML(`
---
foo:
  - <<: (( merge replace bar ))
  - bar:
      alice: alice
      bob: bob
ref: (( foo.[0].alice ))
`)

			stub := parseYAML(`
---
foo:
  - not
  - merged
bar:
  - alice: merged
  - bob: merged
`)

			resolved := parseYAML(`
---
foo:
  - alice: merged
  - bob: merged
ref: merged
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("merging field value", func() {
		It("merges with redirected map, but not with original path", func() {
			source := parseYAML(`
---
foo: (( merge bar ))
`)

			stub := parseYAML(`
---
foo:
  alice: not merged
bar:
  alice: alice
  bob: bob
`)

			resolved := parseYAML(`
---
foo:
  alice: alice
  bob: bob
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges with nothing", func() {
			source := parseYAML(`
---
foo: (( merge nothing || "default" ))
`)

			stub := parseYAML(`
---
foo:
  alice: not merged
`)

			resolved := parseYAML(`
---
foo: default
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

		It("evaluates the node with indirection combined with default", func() {
			source := parseYAML(`
---
meta:
  net: "10.10"

networks:
  some_network:
    type: manual
    subnets:
      - range: (( meta.net ".16.0/20" ))
        name: default_unused
        reserved:
          - (( meta.net ".16.2 - " meta.net ".16.9" ))
          - (( meta.net ".16.255 - " meta.net ".16.255" ))
        static:
          - (( meta.net ".16.10 - " meta.net ".16.254" ))
        gateway: (( meta.net ".16.1" ))
        dns:
          - (( meta.net ".0.2" ))

jobs:
- name: some_job
  resource_pool: some_pool
  instances: 2
  networks:
  - name: some_network
    static_ips: (( static_ips(0, 4) || nil ))
`)

			resolved := parseYAML(`
---
meta:
  net: "10.10"

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

	Describe("map splicing", func() {
		It("merges one map over another", func() {
			source := parseYAML(`
---
properties:
  something:
    foo:
      <<: (( merge ))
      key: a
      val: b
      some:
        s: stuff
        d: blah
`)

			stub := parseYAML(`
---
properties:
  something:
    foo:
      val: c
      some:
        go: home
`)

			resolved := parseYAML(`
---
properties:
  something:
    foo:
      key: a
      val: c
      some:
        s: stuff
        d: blah
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("merges one map over another and resolves inbound references", func() {
			source := parseYAML(`
---
properties:
  something:
    foo:
      <<: (( merge ))
      key: a
      val: b
      some:
        s: stuff
        d: blah
  refkey: (( properties.something.foo.key ))
  refval: (( properties.something.foo.val ))
`)

			stub := parseYAML(`
---
properties:
  something:
    foo:
      val: c
      some:
        go: home
`)

			resolved := parseYAML(`
---
properties:
  something:
    foo:
      key: a
      val: c
      some:
        s: stuff
        d: blah
  refkey: a
  refval: c
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("list splicing", func() {
		It("merges one list into another", func() {
			source := parseYAML(`
---
properties:
  something:
    - a
    - <<: (( list ))
    - b
  list:
    - c
    - d
`)

			resolved := parseYAML(`
---
properties:
  something:
    - a
    - c
    - d
    - b
  list:
    - c
    - d
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("merges stub", func() {
			source := parseYAML(`
---
properties:
  something:
    - a
    - <<: (( merge ))
    - b
`)

			stub := parseYAML(`
---
properties:
  something:
    - c
    - d
`)

			resolved := parseYAML(`
---
properties:
  something:
    - a
    - c
    - d
    - b
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("redirects stub", func() {
			source := parseYAML(`
---
properties:
  something:
    - a
    - <<: (( merge alt ))
    - b
`)

			stub := parseYAML(`
---
properties:
  something:
    - e
    - f
alt:
  - c
  - d
`)

			resolved := parseYAML(`
---
properties:
  something:
    - a
    - c
    - d
    - b
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		Context("when names match", func() {
			It("replaces existing entries with matching names", func() {
				source := parseYAML(`
---
properties:
  something:
    - name: a
      value: 1
    - <<: (( merge ))
    - name: b
      value: 2
`)

				stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

				resolved := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
    - name: b
      value: 2
`)

				Expect(source).To(FlowAs(resolved, stub))
			})

			It("resolves existing entries replaced with matching names", func() {
				source := parseYAML(`
---
properties:
  something:
    - name: a
      value: 1
    - <<: (( merge ))
    - name: b
      value: 2
ref: (( properties.something.[0].value ))
`)

				stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

				resolved := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
    - name: b
      value: 2
ref: 10
`)

				Expect(source).To(FlowAs(resolved, stub))
			})

			It("replaces existing entries with redirected matching names", func() {
				source := parseYAML(`
---
properties:
  something:
    - name: a
      value: 1
    - <<: (( merge alt.something ))
    - name: b
      value: 2
`)

				stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 100
    - name: c
      value: 300

alt:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

				resolved := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
    - name: b
      value: 2
`)

				Expect(source).To(FlowAs(resolved, stub))
			})

			It("resolves existing entries replaced with redirected matching names", func() {
				source := parseYAML(`
---
properties:
  something:
    - name: a
      value: 1
    - <<: (( merge alt.something ))
    - name: b
      value: 2
ref: (( properties.something.a.value ))
`)

				stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 100
    - name: c
      value: 300

alt:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

				resolved := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
    - name: b
      value: 2
ref: 10
`)

				Expect(source).To(FlowAs(resolved, stub))
			})
		})

		It("uses redirected matching names, but not original path", func() {
			source := parseYAML(`
---
properties:
  something: (( merge alt.something ))
`)

			stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 100
    - name: b
      value: 200

alt:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

			resolved := parseYAML(`
---
properties:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

			Expect(source).To(FlowAs(resolved, stub))
		})

		It("avoids override by original path, which occured by traditional redirection", func() {
			source := parseYAML(`
---
alt:
  something: (( merge ))

properties:
  something: (( prefer alt.something ))
`)

			stub := parseYAML(`
---
properties:
  something:
    - name: a
      value: 100
    - name: b
      value: 200

alt:
  something:
    - name: a
      value: 10
    - name: c
      value: 30
`)

			resolved := parseYAML(`
---
alt:
  something:
    - name: a
      value: 10
    - name: c
      value: 30

properties:
  something:
    - name: a
      value: 100
    - name: c
      value: 30
`)

			Expect(source).To(FlowAs(resolved, stub))
		})
	})

	Describe("for list expressions", func() {
		It("evaluates lists", func() {
			source := parseYAML(`
---
foo: (( [ "a", "b" ] ))
`)
			resolved := parseYAML(`
---
foo:
  - a
  - b
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates lists with references", func() {
			source := parseYAML(`
---
a: alice
b: bob
foo: (( [ a, b ] || "failed" ))
`)
			resolved := parseYAML(`
---
a: alice
b: bob
foo:
  - alice
  - bob
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates for lists with deep references", func() {
			source := parseYAML(`
---
a: alice
b: bob
c: (( b ))
foo: (( [ a, c ] || "failed" ))
`)
			resolved := parseYAML(`
---
a: alice
b: bob
c: bob
foo:
  - alice
  - bob
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("failes for lists with unresolved references", func() {
			source := parseYAML(`
---
a: alice
foo: (( [ a, b ] || "failed" ))
`)
			resolved := parseYAML(`
---
a: alice
foo: failed
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("for arithmetic expressions", func() {
		Context("addition", func() {
			It("evaluates addition", func() {
				source := parseYAML(`
---
foo: (( 1 + 2 + 3 ))
`)
				resolved := parseYAML(`
---
foo: 6
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution", func() {
				source := parseYAML(`
---
a: 1
b: 2
c: (( b ))
foo: (( a + c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 1
b: 2
c: 2
foo: 3
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution until failure", func() {
				source := parseYAML(`
---
a: 1
b: 2
foo: (( a + c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 1
b: 2
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("subtraction", func() {
			It("evaluates subtraction", func() {
				source := parseYAML(`
---
foo: (( 6 - 3 - 2 ))
`)
				resolved := parseYAML(`
---
foo: 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution", func() {
				source := parseYAML(`
---
a: 3
b: 2
c: (( b ))
foo: (( a - c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 3
b: 2
c: 2
foo: 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution until failure", func() {
				source := parseYAML(`
---
a: 3
b: 2
foo: (( a - c || "failed" ))
`)

				resolved := parseYAML(`
---
a: 3
b: 2
foo: failed
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("multiplication", func() {
			It("evaluates multiplication", func() {
				source := parseYAML(`
---
foo: (( 6 * 2 * 3 ))
`)
				resolved := parseYAML(`
---
foo: 36
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution", func() {
				source := parseYAML(`
---
a: 6
b: 2
c: (( b ))
foo: (( a * c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 6
b: 2
c: 2
foo: 12
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution until failure", func() {
				source := parseYAML(`
---
a: 6
b: 2
foo: (( a * c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 6
b: 2
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("division", func() {
			It("evaluates division", func() {
				source := parseYAML(`
---
foo: (( 6 / 2 / 3 ))
`)
				resolved := parseYAML(`
---
foo: 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("division by zero fails", func() {
				source := parseYAML(`
---
foo: (( 6 / 0 || "failed" ))
`)
				resolved := parseYAML(`
---
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution", func() {
				source := parseYAML(`
---
a: 6
b: 2
c: (( b ))
foo: (( a / c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 6
b: 2
c: 2
foo: 3
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution until failure", func() {
				source := parseYAML(`
---
a: 6
b: 2
foo: (( a / c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 6
b: 2
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("modulo", func() {
			It("evaluates modulo", func() {
				source := parseYAML(`
---
foo: (( 13 % ( 2 * 3 )))
`)
				resolved := parseYAML(`
---
foo: 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("modulo by zero fails", func() {
				source := parseYAML(`
---
foo: (( 13 % ( 2 - 2 ) || "failed" ))
`)
				resolved := parseYAML(`
---
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution", func() {
				source := parseYAML(`
---
a: 7
b: 2
c: (( b ))
foo: (( a % c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 7
b: 2
c: 2
foo: 1
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates incremental expression resolution until failure", func() {
				source := parseYAML(`
---
a: 7
b: 2
foo: (( a / c || "failed" ))
`)
				resolved := parseYAML(`
---
a: 7
b: 2
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("mixed levels", func() {
			It("evaluates multiplication first", func() {
				source := parseYAML(`
---
foo: (( 6 + 2 * 3 ))
`)
				resolved := parseYAML(`
---
foo: 12
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates addition last", func() {
				source := parseYAML(`
---
foo: (( 6 * 2 + 3 ))
`)
				resolved := parseYAML(`
---
foo: 15
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		It("evaluates arithmetic before concatenation", func() {
			source := parseYAML(`
---
foo: (( "prefix" 6 * 2 + 3 "suffix" ))
`)

			resolved := parseYAML(`
---
foo: prefix15suffix
`)

			Expect(source).To(FlowAs(resolved))
		})

		It("concatenates arithmetic values as string", func() {
			source := parseYAML(`
---
foo: ((  6 * 2 + 3 15 ))
`)

			resolved := parseYAML(`
---
foo: "1515"
`)

			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("for logical expressions", func() {
		It("evaluates not", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( !foo ))
`)
			resolved := parseYAML(`
---
foo: 5
bar: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates and", func() {
			source := parseYAML(`
---
foo: (( 0 ))
bar: (( !foo -and true))
`)
			resolved := parseYAML(`
---
foo: 0
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates or", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( !foo -or true))
`)
			resolved := parseYAML(`
---
foo: 5
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates <=", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo <= 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates <", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo < 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates >=", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo >= 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates >", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo > 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates ==", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo == 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates == of lists", func() {
			source := parseYAML(`
---
foo: 
  - alice
  - bob
bar: (( foo == [ "alice","bob" ] ))
`)
			resolved := parseYAML(`
---
foo: 
  - alice
  - bob
bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates == of lists to false", func() {
			source := parseYAML(`
---
foo: 
  - alice
  - bob
bar: (( foo == [ "alice","paul" ] ))
`)
			resolved := parseYAML(`
---
foo: 
  - alice
  - bob
bar: false
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates == of maps", func() {
			source := parseYAML(`
---
foo: 
  a: 1
  b: 2

comp:
  a: 1
  b: 2

bar: (( foo == comp ))
`)
			resolved := parseYAML(`
---
foo: 
  a: 1
  b: 2

comp:
  a: 1
  b: 2

bar: true
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("evaluates !=", func() {
			source := parseYAML(`
---
foo: (( 5 ))
bar: (( foo != 5))
`)
			resolved := parseYAML(`
---
foo: 5
bar: false
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when concatenating a list", func() {
		Context("with incremental expression resolution", func() {
			It("evaluates in case of successfully completed operand resolution", func() {
				source := parseYAML(`
---
a: alice
b: bob
c: (( b ))
foo: (( a "+" c || "failed" ))
`)
				resolved := parseYAML(`
---
a: alice
b: bob
c: bob
foo: alice+bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("fails only after failed final resolution", func() {
				source := parseYAML(`
---
a: alice
b: bob
foo: (( a "+" c || "failed" ))
`)
				resolved := parseYAML(`
---
a: alice
b: bob
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with other lists", func() {
			It("yields a joined list", func() {
				source := parseYAML(`
---
foo: (( [1,2,3] [ 2 * 3 ] [4,5,6] ))
`)

				resolved := parseYAML(`
---
foo: [1,2,3,6,4,5,6]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with an integer", func() {
			It("appends the value to the list", func() {
				source := parseYAML(`
---
foo: (( [1,2,3] 4 5 ))
`)

				resolved := parseYAML(`
---
foo: [1,2,3,4,5]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with a string", func() {
			It("appends the value to the list", func() {
				source := parseYAML(`
---
foo: (( [1,2,3] "foo" "bar" ))
`)

				resolved := parseYAML(`
---
foo: [1,2,3,"foo","bar"]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with a map", func() {
			It("appends the map to the list", func() {
				source := parseYAML(`
---
bar:
  alice: and bob
foo: (( [1,2,3] bar ))
`)

				resolved := parseYAML(`
---
bar:
  alice: and bob
foo: [1,2,3,{"alice": "and bob"}]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with a nested string concatenation", func() {
			It("appends the value to the list", func() {
				source := parseYAML(`
---
foo: (( [1,2,3] ("foo" "bar") ))
`)

				resolved := parseYAML(`
---
foo: [1,2,3,"foobar"]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("with a nested list concatenation", func() {
			It("joins the list", func() {
				source := parseYAML(`
---
foo: (( [1,2,3] ([] "bar") ))
`)

				resolved := parseYAML(`
---
foo: [1,2,3,"bar"]
`)

				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("when joining", func() {
		It("joins single value", func() {
			source := parseYAML(`
---
foo: (( join( ", ", "alice") ))
`)
			resolved := parseYAML(`
---
foo: alice
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("joins strings and integers", func() {
			source := parseYAML(`
---
foo: (( join( ", ", "alice", "bob", 5) ))
`)
			resolved := parseYAML(`
---
foo: alice, bob, 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("joins elements from lists", func() {
			source := parseYAML(`
---
list:
  - alice
  - bob
foo: (( join( ", ", list, 5) ))
`)
			resolved := parseYAML(`
---
list:
  - alice
  - bob
foo: alice, bob, 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("joins elements from inline list", func() {
			source := parseYAML(`
---
b: bob
foo: (( join( ", ", [ "alice", b ] ) ))
`)
			resolved := parseYAML(`
---
b: bob
foo: alice, bob
`)
			Expect(source).To(FlowAs(resolved))
		})

		Context("with incremental expression resolution", func() {
			It("evaluates in case of successfully completed operand resolution", func() {
				source := parseYAML(`
---
a: alice
b: bob
c: (( b ))
foo: (( join( ", ", a, c) || "failed" ))
`)
				resolved := parseYAML(`
---
a: alice
b: bob
c: bob
foo: alice, bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates in case of successfully completed list operand resolution", func() {
				source := parseYAML(`
---
list:
  - alice
  - (( c ))
b: bob
c: (( b ))
foo: (( join( ", ", list) || "failed" ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - bob
b: bob
c: bob
foo: alice, bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("evaluates in case of successfully completed list expression resolution", func() {
				source := parseYAML(`
---
b: bob
c: (( b ))
foo: (( join( ", ", [ "alice", c ] ) || "failed" ))
`)
				resolved := parseYAML(`
---
b: bob
c: bob
foo: alice, bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("fails only after failed final resolution", func() {
				source := parseYAML(`
---
a: alice
b: bob
foo: (( join( ", ", a, c) || "failed" ))
`)
				resolved := parseYAML(`
---
a: alice
b: bob
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("fails only after failed final list resolution", func() {
				source := parseYAML(`
---
foo: (( join( ", ", [ "alice", c ] ) || "failed" ))
`)
				resolved := parseYAML(`
---
foo: failed
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("when splitting", func() {
		It("splits single value", func() {
			source := parseYAML(`
---
foo: (( split( ",", "alice") ))
`)
			resolved := parseYAML(`
---
foo:
 - alice
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("splits multiple values", func() {
			source := parseYAML(`
---
foo: (( split( ",", "alice,bob") ))
`)
			resolved := parseYAML(`
---
foo:
 - alice
 - bob
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when trimming", func() {
		It("trims strings", func() {
			source := parseYAML(`
---
foo: (( trim( "  alice ") ))
`)
			resolved := parseYAML(`
---
foo: alice
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("trims dedicated characters", func() {
			source := parseYAML(`
---
foo: (( trim( "alice", "ae") ))
`)
			resolved := parseYAML(`
---
foo: lic
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("trims lists", func() {
			source := parseYAML(`
---
foo: (( trim( split(",","alice, bob ")) ))
`)
			resolved := parseYAML(`
---
foo:
  - alice
  - bob
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("calling length", func() {
		It("calculates string length", func() {
			source := parseYAML(`
---
foo: (( length( "alice") ))
`)
			resolved := parseYAML(`
---
foo: 5
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("calculates list length", func() {
			source := parseYAML(`
---
foo: (( length( ["alice","bob"]) ))
`)
			resolved := parseYAML(`
---
foo: 2
`)
			Expect(source).To(FlowAs(resolved))
		})

		It("calculates map length", func() {
			source := parseYAML(`
---
map:
  alice: 25
  bob: 24

foo: (( length( map) ))
`)
			resolved := parseYAML(`
---
map:
  alice: 25
  bob: 24
foo: 2
`)
			Expect(source).To(FlowAs(resolved))
		})
	})

	Describe("when doing a mapping", func() {
		Context("for a list", func() {
			It("maps simple expression", func() {
				source := parseYAML(`
---
list:
  - alice
  - bob
mapped: (( map[list|x|->x] ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - bob
mapped:
  - alice
  - bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps index expression", func() {
				source := parseYAML(`
---
list:
  - alice
  - bob
mapped: (( map[list|y,x|->y x] ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - bob
mapped:
  - 0alice
  - 1bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps concatenation expression", func() {
				source := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
mapped: (( map[list|x|->x ":" port] ))
`)
				resolved := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
mapped:
  - alice:4711
  - bob:4711
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps reference expression", func() {
				source := parseYAML(`
---
list:
  - name: alice
    age: 25
  - name: bob
    age: 24
names: (( map[list|x|->x.name] ))
`)
				resolved := parseYAML(`
---
list:
  - name: alice
    age: 25
  - name: bob
    age: 24
names:
  - alice
  - bob
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps concatenation expression without failure", func() {
				source := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
mapped: (( map[list|x|->x ":" port] || "failed" ))
`)
				resolved := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
mapped:
  - alice:4711
  - bob:4711
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps concatenation expression with failure", func() {
				source := parseYAML(`
---
list:
  - alice
  - bob
mapped: (( map[list|x|->x ":" port] || "failed" ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - bob
mapped: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("works with nested expressions", func() {
				source := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
joined: (( join( ", ", map[list|x|->x ":" port] ) || "failed" ))
`)
				resolved := parseYAML(`
---
port: 4711
list:
  - alice
  - bob
joined: alice:4711, bob:4711
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("works with nested failing expressions", func() {
				source := parseYAML(`
---
list:
  - alice
  - bob
joined: (( join( ", ", map[list|x|->x ":" port] ) || "failed" ))
`)
				resolved := parseYAML(`
---
list:
  - alice
  - bob
joined: failed
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps with referenced expression", func() {
				source := parseYAML(`
---
map: '|x|->x'
list:
  - alice
  - bob
mapped: (( map[list|lambda map] ))
`)
				resolved := parseYAML(`
---
map: '|x|->x'
list:
  - alice
  - bob
mapped:
  - alice
  - bob
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("for a map", func() {
			It("maps simple expression", func() {
				source := parseYAML(`
---
map:
  alice: 25
  bob: 24
mapped: (( map[map|x|->x] ))
`)
				resolved := parseYAML(`
---
map:
  alice: 25
  bob: 24
mapped:
  - 25
  - 24
`)
				Expect(source).To(FlowAs(resolved))
			})

			It("maps key expression", func() {
				source := parseYAML(`
---
map:
  alice: 25
  bob: 24
mapped: (( map[map|y,x|->y x] ))
`)
				resolved := parseYAML(`
---
map:
  alice: 25
  bob: 24
mapped:
  - alice25
  - bob24
`)
				Expect(source).To(FlowAs(resolved))
			})
		})
	})

	Describe("merging lists with specified key", func() {
		Context("no merge", func() {
			It("clean up key tag", func() {
				source := parseYAML(`
---
list:
  - key:address: a
    attr: b
  - address: c
    attr: d
`)
				resolved := parseYAML(`
---
list:
  - address: a
    attr: b
  - address: c
    attr: d
`)
				Expect(source).To(FlowAs(resolved))
			})
		})

		Context("auto merge with key tag", func() {
			It("overrides matching key entries", func() {
				source := parseYAML(`
---
list:
  - key:address: a
    attr: b
  - address: c
    attr: d
`)
				stub := parseYAML(`
---
list:
  - address: c
    attr: stub
  - address: e
    attr: f
`)
				resolved := parseYAML(`
---
list:
  - address: a
    attr: b
  - address: c
    attr: stub
`)
				Expect(source).To(FlowAs(resolved, stub))
			})
		})

		Context("explicit merge with key tag", func() {
			It("overrides matching key entries", func() {
				source := parseYAML(`
---
list:
  - <<: (( merge on address ))
  - address: a
    attr: b
  - address: c
    attr: d
`)
				stub := parseYAML(`
---
list:
  - address: c
    attr: stub
  - address: e
    attr: f
`)
				resolved := parseYAML(`
---
list:
  - address: e
    attr: f
  - address: a
    attr: b
  - address: c
    attr: stub
`)
				Expect(source).To(FlowAs(resolved, stub))
			})
		})
	})
})

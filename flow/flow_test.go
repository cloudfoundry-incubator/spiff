package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/dynaml"
	"github.com/cloudfoundry-incubator/spiff/yaml"
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
			Expect(err).To(Equal(UnresolvedNodes{
				Nodes: []UnresolvedNode{
					{
						Node: yaml.NewNode(
							dynaml.AutoExpr{Path: []string{"foo"}},
							"test",
						),
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
bar:
  alice: alice
  bob: bob
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
  something: (( alt.something ))
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

	
	Describe("for arithmetic expressions", func() {
		///////////////////////
		// Addition
		///////////////////////
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
		
		///////////////////////
		// Subtraction
		///////////////////////
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
		
		///////////////////////
		// Multiplication
		///////////////////////
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
		
		///////////////////////
		// Division
		///////////////////////
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
		
		///////////////////////
		// Modulo
		///////////////////////
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
	
	Describe("concatenation", func() {
		It("of lists", func() {
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
})

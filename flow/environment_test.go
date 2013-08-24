package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/yaml"
)

var _ = Describe("Environment", func() {
	Describe("finding references", func() {
		tree := parseYAML(`
---
foo:
  bar: 42
 `)

		environment := Environment{
			Scope: []map[string]yaml.Node{tree.(map[string]yaml.Node)},
		}

		Context("when the first step is found", func() {
			Context("and the root contains the path", func() {
				It("returns the found node", func() {
					Expect(environment.FindReference([]string{"foo", "bar"})).To(Equal(42))
				})
			})

			Context("and the root does NOT contain the path", func() {
				It("returns nil", func() {
					Expect(environment.FindReference([]string{"foo", "x"})).To(BeNil())
				})
			})
		})

		Context("when the first step is farther away", func() {
			tree := parseYAML(`
---
foo:
  bar:
    baz: x
  fizz:
    buzz: 42
`)

			environment := Environment{
				Scope: []map[string]yaml.Node{
					tree.(map[string]yaml.Node),
					tree.(map[string]yaml.Node)["foo"].(map[string]yaml.Node),
					tree.(map[string]yaml.Node)["foo"].(map[string]yaml.Node)["bar"].(map[string]yaml.Node),
				},
			}

			It("finds the root and the path", func() {
				Expect(environment.FindReference([]string{"fizz", "buzz"})).To(Equal(42))
			})
		})

		Context("when the first step shadows something farther away", func() {
			tree := parseYAML(`
---
foo:
  bar:
    baz: x
    fizz:
      buzz: 123
  fizz:
    buzz: 42
`)

			environment := Environment{
				Scope: []map[string]yaml.Node{
					tree.(map[string]yaml.Node),
					tree.(map[string]yaml.Node)["foo"].(map[string]yaml.Node),
					tree.(map[string]yaml.Node)["foo"].(map[string]yaml.Node)["bar"].(map[string]yaml.Node),
				},
			}

			It("finds the nearest occurrence", func() {
				Expect(environment.FindReference([]string{"fizz", "buzz"})).To(Equal(123))
			})
		})

		Context("when the first step is NOT found", func() {
			It("returns nil", func() {
				Expect(environment.FindReference([]string{"x"})).To(BeNil())
			})
		})
	})

	Describe("finding a path from the root", func() {
		tree := parseYAML(`
---
foo:
  bar:
    baz: found
`)

		environment := Environment{
			Scope: []map[string]yaml.Node{
				tree.(map[string]yaml.Node),
			},
		}

		It("returns the node found by the path from the root", func() {
			Expect(environment.FindFromRoot([]string{"foo", "bar", "baz"})).To(Equal("found"))
		})

		Context("when the path cannot be found", func() {
			It("returns nil", func() {
				Expect(environment.FindFromRoot([]string{"foo", "bar", "biscuit"})).To(BeNil())
			})
		})
	})

	Describe("finding a path in the stubs", func() {
		stub1 := parseYAML(`
---
a: 1
c: 3
`)

		stub2 := parseYAML(`
---
b: 2
c: 4
`)

		environment := Environment{
			Stubs: []yaml.Node{stub1, stub2},
		}

		Context("when the first stub contains the path", func() {
			It("uses the value from the first stub", func() {
				Expect(environment.FindInStubs([]string{"a"})).To(Equal(1))
			})
		})

		Context("when the second stub contains the path", func() {
			It("uses the value from the second stub", func() {
				Expect(environment.FindInStubs([]string{"b"})).To(Equal(2))
			})
		})

		Context("when the both stubs contain the path", func() {
			It("returns the value from the first stub", func() {
				Expect(environment.FindInStubs([]string{"c"})).To(Equal(3))
			})
		})

		Context("when neither stub contains the path", func() {
			It("returns nil", func() {
				Expect(environment.FindInStubs([]string{"d"})).To(BeNil())
			})
		})
	})
})

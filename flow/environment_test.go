package flow

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

var _ = Describe("Environment", func() {
	Describe("finding references", func() {
		tree := parseYAML(`
---
foo:
  bar: 42
 `)

		environment := Environment{
			Scope: []map[string]yaml.Node{tree.Value().(map[string]yaml.Node)},
		}

		Context("when the first step is found", func() {
			Context("and the root contains the path", func() {
				It("returns the found node", func() {
					val, found := environment.FindReference([]string{"foo", "bar"})
					Expect(found).To(BeTrue())
					Expect(val.Value()).To(Equal(int64(42)))
				})
			})

			Context("and the path goes through a list with a named hash", func() {
				tree := parseYAML(`
---
foos:
- name: bar
  baz: 42
`)

				environment := Environment{
					Scope: []map[string]yaml.Node{tree.Value().(map[string]yaml.Node)},
				}

				It("treats the name as the key", func() {
					val, found := environment.FindReference([]string{"foos", "bar", "baz"})
					Expect(found).To(BeTrue())
					Expect(val.Value()).To(Equal(int64(42)))
				})
			})

			Context("and the root does NOT contain the path", func() {
				It("returns false as the second value", func() {
					_, found := environment.FindReference([]string{"foo", "x"})
					Expect(found).To(BeFalse())
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
					tree.Value().(map[string]yaml.Node),
					tree.Value().(map[string]yaml.Node)["foo"].Value().(map[string]yaml.Node),
					tree.Value().(map[string]yaml.Node)["foo"].Value().(map[string]yaml.Node)["bar"].Value().(map[string]yaml.Node),
				},
			}

			It("finds the root and the path", func() {
				val, found := environment.FindReference([]string{"fizz", "buzz"})
				Expect(found).To(BeTrue())
				Expect(val.Value()).To(Equal(int64(42)))
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
					tree.Value().(map[string]yaml.Node),
					tree.Value().(map[string]yaml.Node)["foo"].Value().(map[string]yaml.Node),
					tree.Value().(map[string]yaml.Node)["foo"].Value().(map[string]yaml.Node)["bar"].Value().(map[string]yaml.Node),
				},
			}

			It("finds the nearest occurrence", func() {
				val, found := environment.FindReference([]string{"fizz", "buzz"})
				Expect(found).To(BeTrue())
				Expect(val.Value()).To(Equal(int64(123)))
			})
		})

		Context("when the first step is NOT found", func() {
			It("returns false as the second value", func() {
				_, found := environment.FindReference([]string{"x"})
				Expect(found).To(BeFalse())
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
				tree.Value().(map[string]yaml.Node),
			},
		}

		It("returns the node found by the path from the root", func() {
			val, found := environment.FindFromRoot([]string{"foo", "bar", "baz"})
			Expect(found).To(BeTrue())
			Expect(val.Value()).To(Equal("found"))
		})

		Describe("indexing a list", func() {
			tree := parseYAML(`
---
foo:
  bar:
    - buzz: wrong
    - fizz: right
`)

			environment := Environment{
				Scope: []map[string]yaml.Node{
					tree.Value().(map[string]yaml.Node),
				},
			}

			It("accepts [x] for following through lists", func() {
				val, found := environment.FindFromRoot([]string{"foo", "bar", "[1]", "fizz"})
				Expect(found).To(BeTrue())
				Expect(val.Value()).To(Equal("right"))
			})
		})

		Context("when the path cannot be found", func() {
			It("returns false as the second value", func() {
				_, found := environment.FindFromRoot([]string{"foo", "bar", "biscuit"})
				Expect(found).To(BeFalse())
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
				val, found := environment.FindInStubs([]string{"a"})
				Expect(found).To(BeTrue())
				Expect(val.Value()).To(Equal(int64(1)))
			})
		})

		Context("when the second stub contains the path", func() {
			It("uses the value from the second stub", func() {
				val, found := environment.FindInStubs([]string{"b"})
				Expect(found).To(BeTrue())
				Expect(val.Value()).To(Equal(int64(2)))
			})
		})

		Context("when the both stubs contain the path", func() {
			It("returns the value from the first stub", func() {
				val, found := environment.FindInStubs([]string{"c"})
				Expect(found).To(BeTrue())
				Expect(val.Value()).To(Equal(int64(3)))
			})
		})

		Context("when neither stub contains the path", func() {
			It("returns false as the second argument", func() {
				_, found := environment.FindInStubs([]string{"d"})
				Expect(found).To(BeFalse())
			})
		})
	})
})

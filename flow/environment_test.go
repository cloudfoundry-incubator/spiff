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

	Describe("finding a path in the stubs", func() {})
})

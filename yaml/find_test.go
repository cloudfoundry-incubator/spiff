package yaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("finding paths", func() {
	tree := parseYAML(`
---
foo:
  bar:
    baz: found
`)

	It("returns the node found by the path from the root", func() {
		Expect(Find(tree, "foo", "bar", "baz")).To(Equal("found"))
	})

	Describe("indexing a list", func() {
		tree := parseYAML(`
---
foo:
  bar:
    - buzz: wrong
    - fizz: right
`)

		It("accepts [x] for following through lists", func() {
			Expect(Find(tree, "foo", "bar", "[1]", "fizz")).To(Equal("right"))
		})
	})

	Context("when the path cannot be found", func() {
		It("returns nil", func() {
			Expect(Find(tree, "foo", "bar", "biscuit")).To(BeNil())
		})
	})

	Describe("finding a string", func() {
		tree := parseYAML(`
---
foo: a string
bar: 42
`)

		Context("when the found node is a string", func() {
			It("returns the string and true", func() {
				found, ok := FindString(tree, "foo")
				Expect(ok).To(BeTrue())
				Expect(found).To(Equal("a string"))
			})
		})

		Context("when the found node is NOT a string", func() {
			It("returns false", func() {
				_, ok := FindString(tree, "bar")
				Expect(ok).To(BeFalse())
			})
		})

		Context("when the node is not found", func() {
			It("returns false", func() {
				_, ok := FindString(tree, "baz")
				Expect(ok).To(BeFalse())
			})
		})
	})

	Describe("finding an int", func() {
		tree := parseYAML(`
---
foo: 42
bar: a string
`)

		Context("when the found node is an int", func() {
			It("returns the string and true", func() {
				found, ok := FindInt(tree, "foo")
				Expect(ok).To(BeTrue())
				Expect(found).To(Equal(42))
			})
		})

		Context("when the found node is NOT an int", func() {
			It("returns false", func() {
				_, ok := FindInt(tree, "bar")
				Expect(ok).To(BeFalse())
			})
		})

		Context("when the node is not found", func() {
			It("returns false", func() {
				_, ok := FindInt(tree, "baz")
				Expect(ok).To(BeFalse())
			})
		})
	})
})

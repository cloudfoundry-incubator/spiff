package yaml

import (
	"fmt"
	"math"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("finding paths", func() {

	Describe("Find", func() {

		It("fails when the path is not found", func() {
			tree := parseYAML(`
---
foo: bar
`)

			_, found := Find(tree, "foo", "bar", "biscuit")
			Expect(found).To(BeFalse())
		})

		Context("when the node tree contains maps of maps", func() {
			tree := parseYAML(`
---
foo:
  bar:
    baz: found
`)

			It("accepts string keys to index maps", func() {
				val, found := Find(tree, "foo", "bar", "baz")
				Expect(found).To(BeTrue())
				Expect(val).To(Equal(node("found")))
			})
		})

		Context("when the node tree contains maps of lists of maps", func() {
			tree := parseYAML(`
---
foo:
  bar:
    - buzz: wrong
    - fizz: right
`)

			It("accepts [x] to index lists", func() {
				val, found := Find(tree, "foo", "bar", "[1]", "fizz")
				Expect(found).To(BeTrue())
				Expect(val).To(Equal(node("right")))
			})
		})

	})

	Describe("FindString", func() {
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

	Describe("FindInt", func() {

		Context("when the found node is an int", func() {
			intValue := 64
			tree := parseYAML(fmt.Sprintf(`
---
foo: %d
bar: a string
`, intValue))

			It("returns the value and true", func() {
				found, ok := FindInt(tree, "foo")
				Expect(ok).To(BeTrue())
				Expect(found).To(BeNumerically("==", intValue))
			})
		})

		Context("when the found node is a int64", func() {
			int64Value := int64(math.MaxInt32) + 1
			tree := parseYAML(fmt.Sprintf(`
---
foo: %d
bar: a string
`, int64Value))

			It("returns the value and true", func() {
				found, ok := FindInt(tree, "foo")
				Expect(ok).To(BeTrue())
				Expect(found).To(BeNumerically("==", int64Value))
			})
		})

		Context("when the found node is NOT an int", func() {
			tree := parseYAML(`
---
foo: bar
`)
			It("returns false", func() {
				_, ok := FindInt(tree, "foo")
				Expect(ok).To(BeFalse())
			})
		})

		Context("when the node is not found", func() {
			tree := parseYAML(`
---
foo: bar
`)
			It("returns false", func() {
				_, ok := FindInt(tree, "baz")
				Expect(ok).To(BeFalse())
			})
		})
	})
})

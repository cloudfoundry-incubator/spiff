package yaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parsing YAML", func() {
	Describe("maps", func() {
		It("parses maps as strings mapping to Nodes", func() {
			parsed, err := Parse("test", []byte(`foo: "fizz \"buzz\""`))
			Expect(err).NotTo(HaveOccurred())
			Expect(parsed).To(Equal(node(map[string]Node{"foo": node(`fizz "buzz"`)})))
		})

		It("parses maps with block string values", func() {
			parsesAs("foo: |\n  sup\n  :3", map[string]Node{"foo": node("sup\n:3")})
			parsesAs("foo: >\n  sup\n  :3", map[string]Node{"foo": node("sup :3")})
		})

		Context("when the keys are not strings", func() {
			It("fails", func() {
				_, err := Parse("test", []byte("1: foo"))
				Expect(err).To(Equal(NonStringKeyError{Key: 1}))
			})
		})
	})

	Describe("lists", func() {
		It("parses with Node contents", func() {
			parsesAs("- 1\n- two", []Node{node(1), node("two")})
		})
	})

	Describe("integers", func() {
		It("parses as ints", func() {
			parsesAs("1", 1)
			parsesAs("-1", -1)
		})
	})

	Describe("floats", func() {
		It("parses as float64s", func() {
			parsesAs("1.0", 1.0)
			parsesAs("-1.0", -1.0)
		})
	})

	Describe("booleans", func() {
		It("parses as bools", func() {
			parsesAs("true", true)
			parsesAs("false", false)
		})
	})
})

func parsesAs(source string, expr interface{}) {
	parsed, err := Parse("test", []byte(source))
	Expect(err).NotTo(HaveOccurred())
	Expect(parsed).To(Equal(node(expr)))
}

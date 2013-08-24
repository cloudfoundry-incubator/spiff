package yaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/dynaml"
)

var _ = Describe("Parsing YAML", func() {
	Describe("maps", func() {
		It("parses maps as strings mapping to Nodes", func() {
			parsed, err := Parse([]byte(`foo: "fizz \"buzz\""`))
			Expect(err).NotTo(HaveOccured())
			Expect(parsed).To(Equal(map[string]dynaml.Node{"foo": `fizz "buzz"`}))
		})

		It("parses maps with block string values", func() {
			parsesAs("foo: |\n  sup\n  :3", map[string]dynaml.Node{"foo": "sup\n:3"})
			parsesAs("foo: >\n  sup\n  :3", map[string]dynaml.Node{"foo": "sup :3"})
		})

		Context("when the keys are not strings", func() {
			It("fails", func() {
				_, err := Parse([]byte("1: foo"))
				Expect(err).To(Equal(NonStringKeyError{Key: 1}))
			})
		})
	})

	Describe("lists", func() {
		It("parses with dynaml.Node contents", func() {
			parsesAs("- 1\n- two", []dynaml.Node{1, "two"})
		})
	})

	Describe("integers", func() {
		It("parses as ints", func() {
			parsesAs("1", 1)
			parsesAs("-1", -1)
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
	parsed, err := Parse([]byte(source))
	Expect(err).NotTo(HaveOccured())
	Expect(parsed).To(Equal(expr))
}

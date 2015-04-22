package yaml

import (
	"github.com/cloudfoundry-incubator/candiedyaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Node", func() {

	It("Implements Marshaler", func() {
		subject := NewNode("hello world", "source/path")

		_, ok := subject.(candiedyaml.Marshaler)
		Expect(ok).To(BeTrue())
	})

	Describe("Value", func() {
		It("returns the node value", func() {
			subjectValue := "hello world"
			subject := NewNode(subjectValue, "other/path")

			Expect(subject.Value()).To(Equal(subjectValue))
		})
	})

	Describe("SourceName", func() {
		It("returns the source name", func() {
			subjectValue := "hello world"
			subjectSourceName := "source/path"
			subject := NewNode(subjectValue, subjectSourceName)

			Expect(subject.SourceName()).To(Equal(subjectSourceName))
		})
	})

	Describe("MarshalYAML", func() {
		It("returns an empty string (tag) and the value", func() {
			subjectValue := "hello world"
			subjectSourceName := "source/path"
			subject := NewNode(subjectValue, subjectSourceName)

			tag, value := subject.MarshalYAML()
			Expect(tag).To(Equal(""))
			Expect(value).To(Equal(subjectValue))
		})
	})

	Describe("EquivalentToNode", func() {
		Context("Node value is an string", func() {
			var (
				subjectValue = "hello world"
				subject      = NewNode(subjectValue, "some/path")
			)

			It("returns true if the supplied node's value is an equal string", func() {
				object := NewNode(subjectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeTrue())
			})

			It("returns false if the supplied node's value is an int", func() {
				object := NewNode(int(42), "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeFalse())
			})

		})

		Context("Node value is an int", func() {
			var (
				subjectValue = 123
				subject      = NewNode(subjectValue, "some/path")
			)

			It("returns true if the supplied node's value is an equal int", func() {
				object := NewNode(subjectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeTrue())
			})

			It("returns true if the supplied node's value is an equivelent int64", func() {
				object := NewNode(int64(subjectValue), "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeTrue())
			})

		})

		Context("Node value is an int64", func() {
			var (
				subjectValue = int64(123)
				subject      = NewNode(subjectValue, "some/path")
			)

			It("returns true if the supplied node's value is an equal int64", func() {
				object := NewNode(subjectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeTrue())
			})

			It("returns true if the supplied node's value is an equivelent int", func() {
				object := NewNode(int(subjectValue), "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeTrue())
			})
		})

		Context("Node value is a map", func() {
			var (
				subjectValue = map[string]string{"a": "A", "b": "B"}
				subject      = NewNode(subjectValue, "some/path")
			)

			It("returns true if the supplied node's value is deeply equal", func() {
				objectValue := map[string]string{"b": "B", "a": "A"}
				object := NewNode(objectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeTrue())
			})

			It("returns false if the supplied node's value is unequal", func() {
				objectValue := map[string]string{"a": "A", "c": "C"}
				object := NewNode(objectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeFalse())
			})
		})

		Context("Node value is a slice", func() {
			var (
				subjectValue = []string{"a", "b"}
				subject      = NewNode(subjectValue, "some/path")
			)

			It("returns true if the supplied node's value is deeply equal", func() {
				objectValue := []string{"a", "b"}
				object := NewNode(objectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeTrue())
			})

			It("returns false if the supplied node's value is unequal", func() {
				objectValue := []string{"b", "a"}
				object := NewNode(objectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeFalse())
			})
		})

		Context("Node value is a map[string]Node", func() {
			var (
				subjectValue = map[string]Node{
					"a": NewNode("A", "some/path"),
					"b": NewNode("B", "some/path"),
				}
				subject = NewNode(subjectValue, "some/path")
			)

			It("returns true if the supplied node's value is deeply equal", func() {
				objectValue := map[string]Node{
					"b": NewNode("B", "some/path"),
					"a": NewNode("A", "some/path"),
				}
				object := NewNode(objectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeTrue())
			})

			It("returns false if the supplied node's value is unequal", func() {
				objectValue := map[string]Node{
					"a": NewNode("A", "some/path"),
					"c": NewNode("C", "some/path"),
				}
				object := NewNode(objectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeFalse())
			})
		})

		Context("Node value is a slice of Nodes", func() {
			var (
				subjectValue = []Node{
					NewNode("A", "some/path"),
					NewNode("B", "some/path"),
				}
				subject = NewNode(subjectValue, "some/path")
			)

			It("returns true if the supplied node's value is deeply equal", func() {
				objectValue := []Node{
					NewNode("A", "some/path"),
					NewNode("B", "some/path"),
				}
				object := NewNode(objectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeTrue())
			})

			It("returns false if the supplied node's value is unequal", func() {
				objectValue := []Node{
					NewNode("A", "some/path"),
					NewNode("C", "some/path"),
				}
				object := NewNode(objectValue, "other/path")

				Expect(subject.EquivalentToNode(object)).To(BeFalse())
			})
		})
	})
})

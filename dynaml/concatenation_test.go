package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("concatenation", func() {
	Context("when the left-hand side is a string", func() {
		Context("and the right-hand side is a string", func() {
			It("concatenates both strings", func() {
				expr := ConcatenationExpr{
					StringExpr{"one"},
					StringExpr{"two"},
				}

				Expect(expr).To(EvaluateAs("onetwo", FakeBinding{}))
			})
		})

		Context("and the right-hand side is NOT a string", func() {
			Context("and the right-hand side is a integer", func() {
				It("concatenates both as strings", func() {
					expr := ConcatenationExpr{
						StringExpr{"two"},
						IntegerExpr{42},
					}

					Expect(expr).To(EvaluateAs("two42", FakeBinding{}))
				})
			})

			Context("and the right-hand side is not an integer", func() {
				It("fails", func() {
					expr := ConcatenationExpr{
						StringExpr{"two"},
						ListExpr{[]Expression{StringExpr{"one"}}},
					}

					Expect(expr).To(FailToEvaluate(FakeBinding{}))
				})
			})
		})
	})

	Context("when the left-hand side is a int", func() {
		It("concatenates both as strings", func() {
			expr := ConcatenationExpr{
				IntegerExpr{42},
				StringExpr{"one"},
			}

			Expect(expr).To(EvaluateAs("42one", FakeBinding{}))
		})
	})

	Context("when the left-hand side is a list", func() {
		Context("and the right-hand side is a list", func() {
			It("concatenates both lists", func() {
				expr := ConcatenationExpr{
					ListExpr{[]Expression{StringExpr{"one"}}},
					ListExpr{[]Expression{StringExpr{"two"}}},
				}

				Expect(expr).To(EvaluateAs([]yaml.Node{node("one", nil), node("two", nil)}, FakeBinding{}))
			})
		})

		Context("and the right-hand side is NOT a list", func() {
			Context("and the right-hand side is an integer", func() {
				It("appends to the list", func() {
					expr := ConcatenationExpr{
						ListExpr{[]Expression{StringExpr{"two"}}},
						IntegerExpr{42},
					}

					Expect(expr).To(EvaluateAs([]yaml.Node{node("two", nil), node(42, nil)}, FakeBinding{}))
				})
			})

			Context("and the right-hand side is a string", func() {
				It("appends to the lists", func() {
					expr := ConcatenationExpr{
						ListExpr{[]Expression{StringExpr{"two"}}},
						StringExpr{"one"},
					}

					Expect(expr).To(EvaluateAs([]yaml.Node{node("two", nil), node("one", nil)}, FakeBinding{}))
				})
			})

			Context("and the right-hand side is a map", func() {
				It("appends to the lists", func() {
					expr := ConcatenationExpr{
						ListExpr{[]Expression{StringExpr{"two"}}},
						ReferenceExpr{[]string{"foo"}},
					}

					binding := FakeBinding{
						FoundReferences: map[string]yaml.Node{
							"foo": node(map[string]yaml.Node{"bar": node(42, nil)}, nil),
						},
					}
					Expect(expr).To(EvaluateAs([]yaml.Node{node("two", nil), node(map[string]yaml.Node{"bar": node(42, nil)}, nil)}, binding))
				})
			})

			Context("and the right-hand side is nil", func() {
				It("keeps the lists", func() {
					expr := ConcatenationExpr{
						ListExpr{[]Expression{StringExpr{"two"}}},
						NilExpr{},
					}

					binding := FakeBinding{
						FoundReferences: map[string]yaml.Node{
							"foo": node(map[string]yaml.Node{"bar": node(42, nil)}, nil),
						},
					}
					Expect(expr).To(EvaluateAs([]yaml.Node{node("two", nil)}, binding))
				})
			})
		})
	})
})

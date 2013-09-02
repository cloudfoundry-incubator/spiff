package dynaml

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("parsing", func() {
	Describe("integers", func() {
		It("parses positive numbers", func() {
			parsesAs("1", IntegerExpr{1})
		})

		It("parses negative numbers", func() {
			parsesAs("-1", IntegerExpr{-1})
		})
	})

	Describe("strings", func() {
		It("parses strings with escaped quotes", func() {
			parsesAs(`"foo \"bar\" baz"`, StringExpr{`foo "bar" baz`})
		})
	})

	Describe("nil", func() {
		It("parses nil", func() {
			parsesAs(`nil`, NilExpr{})
		})
	})

	Describe("booleans", func() {
		It("parses true and false", func() {
			parsesAs(`true`, BooleanExpr{true})
			parsesAs(`false`, BooleanExpr{false})
		})
	})

	Describe("merge", func() {
		It("parses as a merge node with the given path", func() {
			parsesAs("merge", MergeExpr{[]string{"foo", "bar"}}, "foo", "bar")
		})
	})

	Describe("auto", func() {
		It("parses as a auto node with the given path", func() {
			parsesAs("auto", AutoExpr{[]string{"foo", "bar"}}, "foo", "bar")
		})
	})

	Describe("references", func() {
		It("parses as a reference node", func() {
			parsesAs("foo.bar.baz", ReferenceExpr{[]string{"foo", "bar", "baz"}})
		})
	})

	Describe("concatenation", func() {
		It("parses adjacent nodes as concatenation", func() {
			parsesAs(
				`"foo" bar`,
				ConcatenationExpr{
					StringExpr{"foo"},
					ReferenceExpr{[]string{"bar"}},
				},
			)

			parsesAs(
				`"foo" bar merge`,
				ConcatenationExpr{
					StringExpr{"foo"},
					ConcatenationExpr{
						ReferenceExpr{[]string{"bar"}},
						MergeExpr{},
					},
				},
			)
		})
	})

	Describe("or", func() {
		It("parses nodes separated by ||", func() {
			parsesAs(
				`"foo" || bar`,
				OrExpr{
					StringExpr{"foo"},
					ReferenceExpr{[]string{"bar"}},
				},
			)

			parsesAs(
				`"foo" || bar || merge`,
				OrExpr{
					StringExpr{"foo"},
					OrExpr{
						ReferenceExpr{[]string{"bar"}},
						MergeExpr{},
					},
				},
			)
		})
	})

	Describe("addition", func() {
		It("parses nodes separated by +", func() {
			parsesAs(
				`"foo" + bar`,
				AdditionExpr{
					StringExpr{"foo"},
					ReferenceExpr{[]string{"bar"}},
				},
			)

			parsesAs(
				`"foo" + bar + merge`,
				AdditionExpr{
					StringExpr{"foo"},
					AdditionExpr{
						ReferenceExpr{[]string{"bar"}},
						MergeExpr{},
					},
				},
			)
		})
	})

	Describe("subtraction", func() {
		It("parses nodes separated by -", func() {
			parsesAs(
				`"foo" - bar`,
				SubtractionExpr{
					StringExpr{"foo"},
					ReferenceExpr{[]string{"bar"}},
				},
			)

			parsesAs(
				`"foo" - bar - merge`,
				SubtractionExpr{
					StringExpr{"foo"},
					SubtractionExpr{
						ReferenceExpr{[]string{"bar"}},
						MergeExpr{},
					},
				},
			)
		})
	})

	Describe("lists", func() {
		It("parses nodes in brackets separated by commas", func() {
			parsesAs(
				`[1, "two", three]`,
				ListExpr{
					[]Expression{
						IntegerExpr{1},
						StringExpr{"two"},
						ReferenceExpr{[]string{"three"}},
					},
				},
			)
		})
	})

	Describe("calls", func() {
		It("parses nodes in arguments to function calls", func() {
			parsesAs(
				`foo(1, "two", three)`,
				CallExpr{
					"foo",
					[]Expression{
						IntegerExpr{1},
						StringExpr{"two"},
						ReferenceExpr{[]string{"three"}},
					},
				},
			)
		})
	})

	Describe("grouping", func() {
		It("influences parser precedence", func() {
			parsesAs(
				`("foo" - bar) - merge`,
				SubtractionExpr{
					SubtractionExpr{
						StringExpr{"foo"},
						ReferenceExpr{[]string{"bar"}},
					},
					MergeExpr{},
				},
			)
		})
	})
})

func parsesAs(source string, expr Expression, path ...string) {
	parsed, err := Parse(source, path)
	Expect(err).NotTo(HaveOccured())
	Expect(parsed).To(Equal(expr))
}

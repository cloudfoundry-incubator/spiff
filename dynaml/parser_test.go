package dynaml

import (
	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = d.Describe("parsing", func() {
	d.Describe("integers", func() {
		d.It("parses positive numbers", func() {
			parsesAs("1", IntegerExpr{1})
		})

		d.It("parses negative numbers", func() {
			parsesAs("-1", IntegerExpr{-1})
		})
	})

	d.Describe("strings", func() {
		d.It("parses strings with escaped quotes", func() {
			parsesAs(`"foo \"bar\" baz"`, StringExpr{`foo "bar" baz`})
		})
	})

	d.Describe("nil", func() {
		d.It("parses nil", func() {
			parsesAs(`nil`, NilExpr{})
		})
	})

	d.Describe("booleans", func() {
		d.It("parses true and false", func() {
			parsesAs(`true`, BooleanExpr{true})
			parsesAs(`false`, BooleanExpr{false})
		})
	})

	d.Describe("merge", func() {
		d.It("parses as a merge node with the given path", func() {
			parsesAs("merge", MergeExpr{[]string{"foo", "bar"}}, "foo", "bar")
		})
	})

	d.Describe("auto", func() {
		d.It("parses as a auto node with the given path", func() {
			parsesAs("auto", AutoExpr{[]string{"foo", "bar"}}, "foo", "bar")
		})
	})

	d.Describe("references", func() {
		d.It("parses as a reference node", func() {
			parsesAs("foo.bar.baz", ReferenceExpr{[]string{"foo", "bar", "baz"}})
		})
	})

	d.Describe("concatenation", func() {
		d.It("parses adjacent nodes as concatenation", func() {
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

	d.Describe("or", func() {
		d.It("parses nodes separated by ||", func() {
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

	d.Describe("addition", func() {
		d.It("parses nodes separated by +", func() {
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

	d.Describe("subtraction", func() {
		d.It("parses nodes separated by -", func() {
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

	d.Describe("lists", func() {
		d.It("parses nodes in brackets separated by commas", func() {
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

	d.Describe("calls", func() {
		d.It("parses nodes in arguments to function calls", func() {
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

	d.Describe("grouping", func() {
		d.It("influences parser precedence", func() {
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

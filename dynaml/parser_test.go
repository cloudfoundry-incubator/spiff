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
			parsesAs("merge alice.bob", MergeExpr{[]string{"alice", "bob"},true,false}, "foo", "bar")
		})
		
		It("parses as a merge node with the environment path", func() {
			parsesAs("merge", MergeExpr{[]string{"foo", "bar"},false,false}, "foo", "bar")
		})
		
		It("parses as a merge replace node with the given path", func() {
			parsesAs("merge replace alice.bob", MergeExpr{[]string{"alice", "bob"},true,true}, "foo", "bar")
		})
		
		It("parses as a merge replace node with the environment path", func() {
			parsesAs("merge replace", MergeExpr{[]string{"foo", "bar"},false,true}, "foo", "bar")
		})
	})

	Describe("auto", func() {
		It("parses as a auto node with the given path", func() {
			parsesAs("auto", AutoExpr{[]string{"foo", "bar"}}, "foo", "bar")
		})
	})

	Describe("references", func() {
		It("parses as a reference node", func() {
			parsesAs("foo.bar-baz.fizz_buzz", ReferenceExpr{[]string{"foo", "bar-baz", "fizz_buzz"}})
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
					ConcatenationExpr{
						StringExpr{"foo"},
						ReferenceExpr{[]string{"bar"}},
					},
					MergeExpr{},
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
					OrExpr{
						StringExpr{"foo"},
						ReferenceExpr{[]string{"bar"}},
					},
					MergeExpr{},
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
					AdditionExpr{
						StringExpr{"foo"},
						ReferenceExpr{[]string{"bar"}},
					},
					MergeExpr{},
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
					SubtractionExpr{
						StringExpr{"foo"},
						ReferenceExpr{[]string{"bar"}},
					},
					MergeExpr{},
				},
			)
		})
	})

	Describe("multiplication", func() {
		It("parses nodes separated by *", func() {
			parsesAs(
				`"foo" * bar`,
				MultiplicationExpr{
					StringExpr{"foo"},
					ReferenceExpr{[]string{"bar"}},
				},
			)

			parsesAs(
				`"foo" * bar * merge`,
				MultiplicationExpr{
					MultiplicationExpr{
						StringExpr{"foo"},
						ReferenceExpr{[]string{"bar"}},
					},
					MergeExpr{},
				},
			)
		})
	})
	
	Describe("division", func() {
		It("parses nodes separated by *", func() {
			parsesAs(
				`"foo" / bar`,
				DivisionExpr{
					StringExpr{"foo"},
					ReferenceExpr{[]string{"bar"}},
				},
			)

			parsesAs(
				`"foo" / bar / merge`,
				DivisionExpr{
					DivisionExpr{
						StringExpr{"foo"},
						ReferenceExpr{[]string{"bar"}},
					},
					MergeExpr{},
				},
			)
		})
	})
	
	Describe("modulo", func() {
		It("parses nodes separated by *", func() {
			parsesAs(
				`"foo" % bar`,
				ModuloExpr{
					StringExpr{"foo"},
					ReferenceExpr{[]string{"bar"}},
				},
			)
		})
	})
	
	Describe("lists", func() {
		It("parses an empty list", func() {
			parsesAs(`[]`, ListExpr{})
		})

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
		
		It("parses nested lists", func() {
			parsesAs(
				`[1, "two", [ three, "four" ] ]`,
				ListExpr{
					[]Expression{
						IntegerExpr{1},
						StringExpr{"two"},
						ListExpr{
							[]Expression{
								ReferenceExpr{[]string{"three"}},
								StringExpr{"four"},
							},
						},
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
		
		It("parses lists in arguments to function calls", func() {
			parsesAs(
				`foo(1, [ "two", three ])`,
				CallExpr{
					"foo",
					[]Expression{
						IntegerExpr{1},
						ListExpr{
							[]Expression{
								StringExpr{"two"},
								ReferenceExpr{[]string{"three"}},
							},
						},
					},
				},
			)
		})
		
		It("parses calls in arguments to function calls", func() {
			parsesAs(
				`foo(1, bar( "two", three ))`,
				CallExpr{
					"foo",
					[]Expression{
						IntegerExpr{1},
						CallExpr{
							"bar", 
							[]Expression{
								StringExpr{"two"},
								ReferenceExpr{[]string{"three"}},
							},
						},
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
	Expect(err).NotTo(HaveOccurred())
	Expect(parsed).To(Equal(expr))
}

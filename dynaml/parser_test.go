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
			parsesAs("merge alice.bob", MergeExpr{[]string{"alice", "bob"}, true, false, true, ""}, "foo", "bar")
		})

		It("parses as a merge node with the environment path", func() {
			parsesAs("merge", MergeExpr{[]string{"foo", "bar"}, false, false, false, ""}, "foo", "bar")
		})

		It("parses as a merge replace node with the given path", func() {
			parsesAs("merge replace alice.bob", MergeExpr{[]string{"alice", "bob"}, true, true, true, ""}, "foo", "bar")
		})

		It("parses as a merge replace node with the environment path", func() {
			parsesAs("merge replace", MergeExpr{[]string{"foo", "bar"}, false, true, true, ""}, "foo", "bar")
		})

		It("parses as a merge require node", func() {
			parsesAs("merge required", MergeExpr{[]string{"foo", "bar"}, false, false, true, ""}, "foo", "bar")
		})

		It("parses as a merge require node", func() {
			parsesAs("merge on key", MergeExpr{[]string{"foo", "bar"}, false, false, false, "key"}, "foo", "bar")
		})

		It("parses as a merge require node", func() {
			parsesAs("merge on key alice.bob", MergeExpr{[]string{"alice", "bob"}, true, false, true, "key"}, "foo", "bar")
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
		It("parses nodes separated by /", func() {
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
		It("parses nodes separated by %", func() {
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
		It("parses simple calls for name", func() {
			parsesAs(
				`foo(1)`,
				CallExpr{
					ReferenceExpr{[]string{"foo"}},
					[]Expression{
						IntegerExpr{1},
					},
				},
			)
		})

		It("parses call for reference", func() {
			parsesAs(
				`foo.bar(1)`,
				CallExpr{
					ReferenceExpr{[]string{"foo", "bar"}},
					[]Expression{
						IntegerExpr{1},
					},
				},
			)
		})

		It("parses call for expression", func() {
			parsesAs(
				`(foo)(1)`,
				CallExpr{
					ReferenceExpr{[]string{"foo"}},
					[]Expression{
						IntegerExpr{1},
					},
				},
			)
		})

		It("parses nodes in arguments to function calls", func() {
			parsesAs(
				`foo(1, "two", three)`,
				CallExpr{
					ReferenceExpr{[]string{"foo"}},
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
					ReferenceExpr{[]string{"foo"}},
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
					ReferenceExpr{[]string{"foo"}},
					[]Expression{
						IntegerExpr{1},
						CallExpr{
							ReferenceExpr{[]string{"bar"}},
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

	Describe("mapping", func() {
		It("parses simple mapping", func() {
			parsesAs(
				`map[list|x|->x]`,
				MapExpr{
					ReferenceExpr{[]string{"list"}},
					LambdaExpr{
						[]string{"x"},
						ReferenceExpr{[]string{"x"}},
					},
				},
			)
		})

		It("parses key/value mapping", func() {
			parsesAs(
				`map[list|x,y|->x]`,
				MapExpr{
					ReferenceExpr{[]string{"list"}},
					LambdaExpr{
						[]string{"x", "y"},
						ReferenceExpr{[]string{"x"}},
					},
				},
			)
		})

		It("parses complex mapping", func() {
			parsesAs(
				`map[list|x|->x ".*"]`,
				MapExpr{
					ReferenceExpr{[]string{"list"}},
					LambdaExpr{
						[]string{"x"},
						ConcatenationExpr{
							ReferenceExpr{[]string{"x"}},
							StringExpr{".*"},
						},
					},
				},
			)
		})

		It("parses mapping expression", func() {
			parsesAs(
				`map[list|mappings.a]`,
				MapExpr{
					ReferenceExpr{[]string{"list"}},
					ReferenceExpr{
						[]string{"mappings", "a"},
					},
				},
			)
		})

		It("parses complex mapping expression", func() {
			parsesAs(
				`map[list|lambda |x|->x ".*"]`,
				MapExpr{
					ReferenceExpr{[]string{"list"}},
					LambdaExpr{
						[]string{"x"},
						ConcatenationExpr{
							ReferenceExpr{[]string{"x"}},
							StringExpr{".*"},
						},
					},
				},
			)
		})
	})

	Describe("lambda expressions", func() {
		It("parses expression with one parameter", func() {
			parsesAs(
				`lambda|x|->x`,
				LambdaExpr{
					[]string{"x"},
					ReferenceExpr{[]string{"x"}},
				},
			)
		})

		It("parses expression with two parameter", func() {
			parsesAs(
				`lambda|x,y|->x / y`,
				LambdaExpr{
					[]string{"x", "y"},
					DivisionExpr{
						ReferenceExpr{[]string{"x"}},
						ReferenceExpr{[]string{"y"}},
					},
				},
			)
		})

		It("parses calculated expression", func() {
			parsesAs(
				`lambda "|x|->x+" ref`,
				LambdaRefExpr{
					Source: ConcatenationExpr{
						StringExpr{"|x|->x+"},
						ReferenceExpr{[]string{"ref"}},
					},
					Path:     []string{"foo", "bar"},
					StubPath: []string{"foo", "bar"},
				},
				"foo", "bar",
			)
		})
	})

	Describe("chained calls and references", func() {
		It("parses reference based chains", func() {
			parsesAs(
				`a.b(1)(2).c(3).e.f(4).g`,
				QualifiedExpr{
					CallExpr{
						QualifiedExpr{
							CallExpr{
								QualifiedExpr{
									CallExpr{
										CallExpr{
											ReferenceExpr{[]string{"a", "b"}},
											[]Expression{
												IntegerExpr{1},
											},
										},
										[]Expression{
											IntegerExpr{2},
										},
									},
									ReferenceExpr{[]string{"c"}},
								},
								[]Expression{
									IntegerExpr{3},
								},
							},
							ReferenceExpr{[]string{"e", "f"}},
						},
						[]Expression{
							IntegerExpr{4},
						},
					},
					ReferenceExpr{[]string{"g"}},
				},
			)
		})

		It("parses list based chains", func() {
			parsesAs(
				`[1,2].a(1)(2).c(3).e.f(4).g`,
				QualifiedExpr{
					CallExpr{
						QualifiedExpr{
							CallExpr{
								QualifiedExpr{
									CallExpr{
										CallExpr{
											QualifiedExpr{
												ListExpr{
													[]Expression{
														IntegerExpr{1},
														IntegerExpr{2},
													},
												},
												ReferenceExpr{[]string{"a"}},
											},
											[]Expression{
												IntegerExpr{1},
											},
										},
										[]Expression{
											IntegerExpr{2},
										},
									},
									ReferenceExpr{[]string{"c"}},
								},
								[]Expression{
									IntegerExpr{3},
								},
							},
							ReferenceExpr{[]string{"e", "f"}},
						},
						[]Expression{
							IntegerExpr{4},
						},
					},
					ReferenceExpr{[]string{"g"}},
				},
			)
		})
	})
})

func parsesAs(source string, expr Expression, path ...string) {
	parsed, err := Parse(source, path, path)
	Expect(err).NotTo(HaveOccurred())
	Expect(parsed).To(Equal(expr))
}

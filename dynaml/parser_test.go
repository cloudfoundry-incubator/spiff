package dynaml

import (
	. "launchpad.net/gocheck"
)

type DynamlParserSuite struct{}

func init() {
	Suite(&DynamlParserSuite{})
}

func parsesAs(c *C, source string, expr Expression, path ...string) {
	parsed, err := Parse(source, path)
	if err != nil {
		c.Error(err)
		return
	}

	c.Assert(parsed, FitsTypeOf, expr)
	c.Assert(parsed, DeepEquals, expr)
}

func (s *DynamlParserSuite) TestParsingNegativeIntegers(c *C) {
	parsesAs(c, "1", IntegerExpr{1})
	parsesAs(c, "-1", IntegerExpr{-1})
}

func (s *DynamlParserSuite) TestParsingStrings(c *C) {
	parsesAs(c, `"foo \"bar\" baz"`, StringExpr{`foo "bar" baz`})
}

func (s *DynamlParserSuite) TestParsingBooleans(c *C) {
	parsesAs(c, "true", BooleanExpr{true})
	parsesAs(c, "false", BooleanExpr{false})
}

func (s *DynamlParserSuite) TestParsingMerge(c *C) {
	parsesAs(c, "merge", MergeExpr{[]string{"foo", "bar"}}, "foo", "bar")
}

func (s *DynamlParserSuite) TestParsingAuto(c *C) {
	parsesAs(c, "auto", AutoExpr{[]string{"foo", "bar"}}, "foo", "bar")
}

func (s *DynamlParserSuite) TestParsingReferences(c *C) {
	parsesAs(c, "foo", ReferenceExpr{[]string{"foo"}})
	parsesAs(c, "foo.bar.baz", ReferenceExpr{[]string{"foo", "bar", "baz"}})
}

func (s *DynamlParserSuite) TestParsingConcatination(c *C) {
	parsesAs(
		c,
		`"foo" bar`,
		ConcatenationExpr{
			StringExpr{"foo"},
			ReferenceExpr{[]string{"bar"}},
		},
	)

	parsesAs(
		c,
		`"foo" bar merge`,
		ConcatenationExpr{
			StringExpr{"foo"},
			ConcatenationExpr{
				ReferenceExpr{[]string{"bar"}},
				MergeExpr{},
			},
		},
	)
}

func (s *DynamlParserSuite) TestParsingOr(c *C) {
	parsesAs(
		c,
		`"foo" || bar`,
		OrExpr{
			StringExpr{"foo"},
			ReferenceExpr{[]string{"bar"}},
		},
	)

	parsesAs(
		c,
		`"foo" || bar || merge`,
		OrExpr{
			StringExpr{"foo"},
			OrExpr{
				ReferenceExpr{[]string{"bar"}},
				MergeExpr{},
			},
		},
	)
}

func (s *DynamlParserSuite) TestParsingAddition(c *C) {
	parsesAs(
		c,
		`"foo" + bar`,
		AdditionExpr{
			StringExpr{"foo"},
			ReferenceExpr{[]string{"bar"}},
		},
	)

	parsesAs(
		c,
		`"foo" + bar + merge`,
		AdditionExpr{
			StringExpr{"foo"},
			AdditionExpr{
				ReferenceExpr{[]string{"bar"}},
				MergeExpr{},
			},
		},
	)
}

func (s *DynamlParserSuite) TestParsingSubtraction(c *C) {
	parsesAs(
		c,
		`"foo" - bar`,
		SubtractionExpr{
			StringExpr{"foo"},
			ReferenceExpr{[]string{"bar"}},
		},
	)

	parsesAs(
		c,
		`"foo" - bar - merge`,
		SubtractionExpr{
			StringExpr{"foo"},
			SubtractionExpr{
				ReferenceExpr{[]string{"bar"}},
				MergeExpr{},
			},
		},
	)
}

func (s *DynamlParserSuite) TestParsingLists(c *C) {
	parsesAs(
		c,
		`[1, "two", three]`,
		ListExpr{
			[]Expression{
				IntegerExpr{1},
				StringExpr{"two"},
				ReferenceExpr{[]string{"three"}},
			},
		},
	)
}

func (s *DynamlParserSuite) TestParsingCalls(c *C) {
	parsesAs(
		c,
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
}

func (s *DynamlParserSuite) TestParsingGrouped(c *C) {
	parsesAs(
		c,
		`("foo" - bar) - merge`,
		SubtractionExpr{
			SubtractionExpr{
				StringExpr{"foo"},
				ReferenceExpr{[]string{"bar"}},
			},
			MergeExpr{},
		},
	)
}

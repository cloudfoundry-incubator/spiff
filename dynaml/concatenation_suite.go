package dynaml

import (
	. "launchpad.net/gocheck"
)

type ConcatenationSuite struct{}

func init() {
	Suite(&ConcatenationSuite{})
}

func (s *ConcatenationSuite) TestConcatenationEvaluate(c *C) {
	expr := ConcatenationExpr{
		StringExpr{"one"},
		StringExpr{"two"},
	}

	c.Assert(expr.Evaluate(FakeContext{}), Equals, "onetwo")
}

func (s *ConcatenationSuite) TestConcatenationEvaluateWithNonStringLHS(c *C) {
	expr := ConcatenationExpr{
		StringExpr{"one"},
		IntegerExpr{42},
	}

	c.Assert(expr.Evaluate(FakeContext{}), IsNil)
}

func (s *ConcatenationSuite) TestConcatenationEvaluateWithNonStringRHS(c *C) {
	expr := ConcatenationExpr{
		IntegerExpr{42},
		StringExpr{"two"},
	}

	c.Assert(expr.Evaluate(FakeContext{}), IsNil)
}

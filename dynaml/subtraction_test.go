package dynaml

import (
	. "launchpad.net/gocheck"
)

type SubtractionSuite struct{}

func init() {
	Suite(&SubtractionSuite{})
}

func (s *SubtractionSuite) TestSubtractionEvaluate(c *C) {
	expr := SubtractionExpr{
		IntegerExpr{6},
		IntegerExpr{2},
	}

	c.Assert(expr.Evaluate(FakeContext{}), Equals, 4)
}

func (s *SubtractionSuite) TestSubtractionEvaluateWithNonIntegerLHS(c *C) {
	expr := SubtractionExpr{
		StringExpr{"lol"},
		IntegerExpr{2},
	}

	c.Assert(expr.Evaluate(FakeContext{}), IsNil)
}

func (s *SubtractionSuite) TestSubtractionEvaluateWithNonIntegerRHS(c *C) {
	expr := SubtractionExpr{
		IntegerExpr{2},
		StringExpr{"lol"},
	}

	c.Assert(expr.Evaluate(FakeContext{}), IsNil)
}

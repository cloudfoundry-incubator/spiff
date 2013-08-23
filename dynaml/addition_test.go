package dynaml

import (
	. "launchpad.net/gocheck"
)

type AdditionSuite struct{}

func init() {
	Suite(&AdditionSuite{})
}

func (s *AdditionSuite) TestAdditionEvaluate(c *C) {
	expr := AdditionExpr{
		IntegerExpr{2},
		IntegerExpr{3},
	}

	c.Assert(expr.Evaluate(FakeContext{}), Equals, 5)
}

func (s *AdditionSuite) TestAdditionEvaluateWithNonIntegerLHS(c *C) {
	expr := AdditionExpr{
		StringExpr{"lol"},
		IntegerExpr{2},
	}

	c.Assert(expr.Evaluate(FakeContext{}), IsNil)
}

func (s *AdditionSuite) TestAdditionEvaluateWithNonIntegerRHS(c *C) {
	expr := AdditionExpr{
		IntegerExpr{2},
		StringExpr{"lol"},
	}

	c.Assert(expr.Evaluate(FakeContext{}), IsNil)
}

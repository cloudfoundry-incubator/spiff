package dynaml

import (
	. "launchpad.net/gocheck"
)

type OrSuite struct{}

func init() {
	Suite(&OrSuite{})
}

func (s *OrSuite) TestOrEvaluate(c *C) {
	expr := OrExpr{
		IntegerExpr{2},
		IntegerExpr{3},
	}

	c.Assert(expr.Evaluate(FakeContext{}), Equals, 2)
}

func (s *OrSuite) TestOrEvaluateWithNilLHS(c *C) {
	expr := OrExpr{
		ReferenceExpr{},
		IntegerExpr{3},
	}

	c.Assert(expr.Evaluate(FakeContext{}), Equals, 3)
}

func (s *OrSuite) TestOrEvaluateWithNilRHS(c *C) {
	expr := OrExpr{
		IntegerExpr{2},
		ReferenceExpr{},
	}

	c.Assert(expr.Evaluate(FakeContext{}), Equals, 2)
}

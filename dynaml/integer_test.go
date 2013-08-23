package dynaml

import (
	. "launchpad.net/gocheck"
)

type IntegerSuite struct{}

func init() {
	Suite(&IntegerSuite{})
}

func (s *IntegerSuite) TestIntegerEvaluate(c *C) {
	expr := IntegerExpr{42}
	c.Assert(expr.Evaluate(FakeContext{}), Equals, 42)
}

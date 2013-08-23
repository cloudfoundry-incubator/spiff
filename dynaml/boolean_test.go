package dynaml

import (
	. "launchpad.net/gocheck"
)

type BooleanSuite struct{}

func init() {
	Suite(&BooleanSuite{})
}

func (s *BooleanSuite) TestBooleanEvaluate(c *C) {
	c.Assert(BooleanExpr{false}.Evaluate(FakeContext{}), Equals, false)
	c.Assert(BooleanExpr{true}.Evaluate(FakeContext{}), Equals, true)
}

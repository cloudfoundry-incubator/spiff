package dynaml

import (
	. "launchpad.net/gocheck"
)

type StringSuite struct{}

func init() {
	Suite(&StringSuite{})
}

func (s *StringSuite) TestStringEvaluate(c *C) {
	c.Assert(StringExpr{"sup"}.Evaluate(FakeContext{}), Equals, "sup")
}

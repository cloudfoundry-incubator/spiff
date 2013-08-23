package dynaml

import (
	. "launchpad.net/gocheck"
)

type ListSuite struct{}

func init() {
	Suite(&ListSuite{})
}

func (s *ListSuite) TestListEvaluate(c *C) {
	expr := ListExpr{
		[]Expression{
			IntegerExpr{1},
			StringExpr{"two"},
		},
	}

	c.Assert(expr.Evaluate(FakeContext{}), DeepEquals, []Node{1, "two"})
}

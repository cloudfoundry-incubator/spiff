package yaml

import (
	. "launchpad.net/gocheck"
)

type YAMLParserSuite struct{}

func init() {
	Suite(&YAMLParserSuite{})
}

func parsesAs(c *C, source string, expr interface{}) {
	parsed, err := Parse([]byte(source))
	if err != nil {
		c.Error(err)
		return
	}

	c.Assert(parsed, FitsTypeOf, expr)
	c.Assert(parsed, DeepEquals, expr)
}

func (s *YAMLParserSuite) TestParsingMaps(c *C) {
	parsesAs(c, `foo: "fizz \"buzz\""`, map[string]Node{"foo": `fizz "buzz"`})
}

func (s *YAMLParserSuite) TestParsingMapsWithBlockStrings(c *C) {
	parsesAs(c, "foo: |\n  sup\n  :3", map[string]Node{"foo": "sup\n:3"})
	parsesAs(c, "foo: >\n  sup\n  :3", map[string]Node{"foo": "sup :3"})
}

func (s *YAMLParserSuite) TestParsingMapsWithNonStringKeysFails(c *C) {
	_, err := Parse([]byte("1: foo"))
	c.Assert(err, FitsTypeOf, NonStringKeyError{})
}

func (s *YAMLParserSuite) TestParsingLists(c *C) {
	parsesAs(c, "- 1\n- two", []Node{1, "two"})
}

func (s *YAMLParserSuite) TestParsingIntegers(c *C) {
	parsesAs(c, "1", 1)
	parsesAs(c, "-1", -1)
}

func (s *YAMLParserSuite) TestParsingBooleans(c *C) {
	parsesAs(c, "true", true)
	parsesAs(c, "false", false)
}

package dynaml

import (
	"strings"

	"github.com/vito/spiff/yaml"
)

type FakeContext struct {
	FoundFromRoot   map[string]yaml.Node
	FoundReferences map[string]yaml.Node
	FoundInStubs    map[string]yaml.Node
}

func (c FakeContext) FindFromRoot(path []string) (yaml.Node, bool) {
	val, found := c.FoundFromRoot[strings.Join(path, ".")]
	return val, found
}

func (c FakeContext) FindReference(path []string) (yaml.Node, bool) {
	val, found := c.FoundReferences[strings.Join(path, ".")]
	return val, found
}

func (c FakeContext) FindInStubs(path []string) (yaml.Node, bool) {
	val, found := c.FoundInStubs[strings.Join(path, ".")]
	return val, found
}

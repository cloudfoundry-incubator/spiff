package dynaml

import (
	"strings"

	"github.com/vito/spiff/yaml"
)

type FakeBinding struct {
	FoundFromRoot   map[string]yaml.Node
	FoundReferences map[string]yaml.Node
	FoundInStubs    map[string]yaml.Node
}

func (c FakeBinding) FindFromRoot(path []string) (yaml.Node, bool) {
	val, found := c.FoundFromRoot[strings.Join(path, ".")]
	return val, found
}

func (c FakeBinding) FindReference(path []string) (yaml.Node, bool) {
	val, found := c.FoundReferences[strings.Join(path, ".")]
	return val, found
}

func (c FakeBinding) FindInStubs(path []string) (yaml.Node, bool) {
	val, found := c.FoundInStubs[strings.Join(path, ".")]
	return val, found
}

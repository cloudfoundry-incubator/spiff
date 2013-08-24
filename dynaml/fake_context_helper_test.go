package dynaml

import (
	"strings"

	"github.com/vito/spiff/yaml"
)

type FakeContext struct {
	FoundReferences map[string]yaml.Node
	FoundInStubs    map[string]yaml.Node
}

func (c FakeContext) FindReference(path []string) yaml.Node {
	return c.FoundReferences[strings.Join(path, ".")]
}

func (c FakeContext) FindInStubs(path []string) yaml.Node {
	return c.FoundInStubs[strings.Join(path, ".")]
}

package dynaml

import (
	"strings"
)

type FakeContext struct {
	FoundReferences map[string]Node
	FoundInStubs    map[string]Node
}

func (c FakeContext) FindReference(path []string) Node {
	return c.FoundReferences[strings.Join(path, ".")]
}

func (c FakeContext) FindInStubs(path []string) Node {
	return c.FoundInStubs[strings.Join(path, ".")]
}

package dynaml

import (
	"strings"
)

type FakeContext struct {
	FoundReferences map[string]Node
}

func (c FakeContext) FindReference(path []string) Node {
	return c.FoundReferences[strings.Join(path, ".")]
}

func (FakeContext) FindInStubs([]string) Node { return nil }

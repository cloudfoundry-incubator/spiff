package dynaml

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type FakeBinding struct {
	FoundFromRoot   map[string]yaml.Node
	FoundReferences map[string]yaml.Node
	FoundInStubs    map[string]yaml.Node

	path     []string
	stubPath []string
}

func (c FakeBinding) Path() []string {
	return c.path
}

func (c FakeBinding) StubPath() []string {
	return c.stubPath
}

func (c FakeBinding) RedirectOverwrite([]string) Binding {
	return c
}

func (c FakeBinding) WithScope(map[string]yaml.Node) Binding {
	return c
}

func (c FakeBinding) WithPath(step string) Binding {
	newPath := make([]string, len(c.path))
	copy(newPath, c.path)
	c.path = append(newPath, step)

	newPath = make([]string, len(c.stubPath))
	copy(newPath, c.stubPath)
	c.stubPath = append(newPath, step)
	return c
}

func (c FakeBinding) GetLocalBinding() map[string]yaml.Node {
	return map[string]yaml.Node{}
}

func (c FakeBinding) FindFromRoot(path []string) (yaml.Node, bool) {
	p := strings.Join(path, ".")
	if len(path) == 0 {
		p = ""
	}
	val, found := c.FoundFromRoot[p]
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

func (c FakeBinding) Flow(source yaml.Node, shouldOverride bool) (yaml.Node, error) {
	return nil, fmt.Errorf("not implemented")
}

package flow

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type Scope []map[string]yaml.Node

type Environment struct {
	Scope Scope
	Path  []string

	Stubs []yaml.Node

	StubPath []string
}

func (e Environment) FindFromRoot(path []string) (yaml.Node, bool) {
	if len(e.Scope) == 0 {
		return nil, false
	}

	return yaml.Find(yaml.NewNode(e.Scope[0], "scope"), path...)
}

func (e Environment) FindReference(path []string) (yaml.Node, bool) {
	root, found := resolveSymbol(path[0], e.Scope)
	if !found {
		return nil, false
	}

	return yaml.Find(root, path[1:]...)
}

func (e Environment) FindInStubs(path []string) (yaml.Node, bool) {
	for _, stub := range e.Stubs {
		val, found := yaml.Find(stub, path...)
		if found {
			return val, true
		}
	}

	return nil, false
}

func (e Environment) WithScope(step map[string]yaml.Node) Environment {
	newScope := make([]map[string]yaml.Node, len(e.Scope))
	copy(newScope, e.Scope)
	e.Scope = append(newScope, step)
	return e
}

func (e Environment) WithPath(step string) Environment {
	newPath := make([]string, len(e.Path))
	copy(newPath, e.Path)
	e.Path = append(newPath, step)

	newPath = make([]string, len(e.StubPath))
	copy(newPath, e.StubPath)
	e.StubPath = append(newPath, step)
	return e
}

func (e Environment) RedirectOverwrite(path []string) Environment {
	e.StubPath = path
	return e
}

func resolveSymbol(name string, context Scope) (yaml.Node, bool) {
	for i := len(context); i > 0; i-- {
		ctx := context[i-1]
		val := ctx[name]
		if val != nil {
			return val, true
		}
	}

	return nil, false
}

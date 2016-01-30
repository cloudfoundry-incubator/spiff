package flow

import (
	"reflect"

	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/dynaml"
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type Scope struct {
	local map[string]yaml.Node
	next  *Scope
	root  *Scope
}

func newScope(outer *Scope, local map[string]yaml.Node) *Scope {
	scope := &Scope{local, outer, nil}
	if outer == nil {
		scope.root = scope
	} else {
		scope.root = outer.root
	}
	return scope
}

type DefaultEnvironment struct {
	scope *Scope
	path  []string

	stubs []yaml.Node

	stubPath []string
}

func (e DefaultEnvironment) Path() []string {
	return e.path
}

func (e DefaultEnvironment) StubPath() []string {
	return e.stubPath
}

func (e DefaultEnvironment) GetLocalBinding() map[string]yaml.Node {
	return map[string]yaml.Node{}
}

func (e DefaultEnvironment) FindFromRoot(path []string) (yaml.Node, bool) {
	if e.scope == nil {
		return nil, false
	}

	return yaml.FindR(true, yaml.NewNode(e.scope.root.local, "scope"), path...)
}

func (e DefaultEnvironment) FindReference(path []string) (yaml.Node, bool) {
	root, found := resolveSymbol(path[0], e.scope)
	if !found {
		return nil, false
	}

	return yaml.FindR(true, root, path[1:]...)
}

func (e DefaultEnvironment) FindInStubs(path []string) (yaml.Node, bool) {
	for _, stub := range e.stubs {
		val, found := yaml.Find(stub, path...)
		if found {
			return val, true
		}
	}

	return nil, false
}

func (e DefaultEnvironment) WithScope(step map[string]yaml.Node) dynaml.Binding {
	e.scope = newScope(e.scope, step)
	return e
}

func (e DefaultEnvironment) WithPath(step string) dynaml.Binding {
	newPath := make([]string, len(e.path))
	copy(newPath, e.path)
	e.path = append(newPath, step)

	newPath = make([]string, len(e.stubPath))
	copy(newPath, e.stubPath)
	e.stubPath = append(newPath, step)
	return e
}

func (e DefaultEnvironment) RedirectOverwrite(path []string) dynaml.Binding {
	e.stubPath = path
	return e
}

func (e DefaultEnvironment) Flow(source yaml.Node, shouldOverride bool) (yaml.Node, dynaml.Status) {
	result := source

	for {
		debug.Debug("@@@ loop:  %+v\n", result)
		next := flow(result, e, shouldOverride)
		debug.Debug("@@@ --->   %+v\n", next)

		if reflect.DeepEqual(result, next) {
			break
		}

		result = next
	}
	debug.Debug("@@@ Done\n")
	unresolved := dynaml.FindUnresolvedNodes(result)
	if len(unresolved) > 0 {
		return nil, dynaml.UnresolvedNodes{unresolved}
	}

	return result, nil
}

func NewEnvironment(stubs []yaml.Node) dynaml.Binding {
	return DefaultEnvironment{stubs: stubs}
}

func resolveSymbol(name string, scope *Scope) (yaml.Node, bool) {
	for scope != nil {
		val := scope.local[name]
		if val != nil {
			return val, true
		}
		scope = scope.next
	}

	return nil, false
}

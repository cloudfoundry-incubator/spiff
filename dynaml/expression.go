package dynaml

import (
	"github.com/shutej/spiff/yaml"
)

type Binding interface {
	ProvidesPhases(StringSet) bool
	Builtin(name string) (Builtin, bool)

	FindFromRoot([]string) (yaml.Node, bool)
	FindReference([]string) (yaml.Node, bool)
	FindInStubs([]string) (yaml.Node, bool)
}

type Expression interface {
	RequiresPhases(Binding) StringSet

	Evaluate(Binding) (yaml.Node, bool)
}

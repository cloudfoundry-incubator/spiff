package dynaml

import (
	"github.com/shutej/spiff/yaml"
)

type Binding interface {
	FindFromRoot([]string) (yaml.Node, bool)
	FindReference([]string) (yaml.Node, bool)
	FindInStubs([]string) (yaml.Node, bool)
	ProvidesPhases(StringSet) bool
}

type Expression interface {
	RequiresPhases() StringSet
	Evaluate(Binding) (yaml.Node, bool)
}

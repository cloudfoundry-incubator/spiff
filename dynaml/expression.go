package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type Context interface {
	FindFromRoot([]string) yaml.Node
	FindReference([]string) yaml.Node
	FindInStubs([]string) yaml.Node
}

type Expression interface {
	Evaluate(Context) yaml.Node
}

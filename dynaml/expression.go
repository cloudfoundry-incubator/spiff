package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type Context interface {
	FindReference([]string) yaml.Node
	FindInStubs([]string) yaml.Node
}

type Expression interface {
	Evaluate(Context) yaml.Node
}

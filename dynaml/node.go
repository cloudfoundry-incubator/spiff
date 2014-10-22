package dynaml

import (
	"github.com/shutej/spiff/yaml"
)

func node(val interface{}) yaml.Node {
	return yaml.NewNode(val, "dynaml")
}

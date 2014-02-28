package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func node(val interface{}) yaml.Node {
	return yaml.NewNode(val, "dynaml")
}

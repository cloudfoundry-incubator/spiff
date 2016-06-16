package dynaml

import (
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func value(node yaml.Node) interface{} {
	if node == nil {
		return nil
	}
	return node.Value()
}

func node(val interface{}, src SourceProvider) yaml.Node {
	source := ""

	if src == nil || len(src.SourceName()) == 0 {
		source = "dynaml"
	} else {
		if !strings.HasPrefix(src.SourceName(), "dynaml@") {
			source = "dynaml@" + src.SourceName()
		} else {
			source = src.SourceName()
		}
	}

	return yaml.NewNode(val, source)
}

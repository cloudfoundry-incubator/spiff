package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type BooleanExpr struct {
	Value bool
}

func (e BooleanExpr) Evaluate(Binding) (yaml.Node, bool) {
	return node(e.Value), true
}

func (e BooleanExpr) String() string {
	return fmt.Sprintf("%v", e.Value)
}

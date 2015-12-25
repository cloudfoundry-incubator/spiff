package dynaml

import (
	"strings"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type ReferenceExpr struct {
	Path []string
}

func (e ReferenceExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	var step yaml.Node
	var ok bool

    info := DefaultInfo()
	fromRoot := e.Path[0] == ""

	for i := 0; i < len(e.Path); i++ {
		if fromRoot {
			step, ok = binding.FindFromRoot(e.Path[1 : i+1])
		} else {
			step, ok = binding.FindReference(e.Path[:i+1])
		}

		if !ok {
			return nil, info, false
		}

		switch step.Value().(type) {
		case Expression:
			return node(e), info, true
		}
	}

	return step, info, true
}

func (e ReferenceExpr) String() string {
	return strings.Join(e.Path, ".")
}

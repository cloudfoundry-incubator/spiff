package dynaml

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry-incubator/spiff/debug"
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

	debug.Debug("reference: %v\n", e.Path)
	for i := 0; i < len(e.Path); i++ {
		if fromRoot {
			step, ok = binding.FindFromRoot(e.Path[1 : i+1])
		} else {
			step, ok = binding.FindReference(e.Path[:i+1])
		}

		debug.Debug("  %d: %v %+v\n", i, ok, step)
		if !ok {
			info.Issue = fmt.Sprintf("'%s' not found", strings.Join(e.Path, "."))
			return nil, info, false
		}

		if !isLocallyResolved(step) {
			debug.Debug("  unresolved\n")
			return node(e), info, true
		}
	}

	if !isResolved(step) {
		debug.Debug("  unresolved\n")
		return node(e), info, true
	}

	debug.Debug("reference %v -> %+v\n", e.Path, step)
	return yaml.ReferencedNode(step), info, true
}

func (e ReferenceExpr) String() string {
	return strings.Join(e.Path, ".")
}

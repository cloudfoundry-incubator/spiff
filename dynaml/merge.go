package dynaml

import (
	"strings"
	"github.com/cloudfoundry-incubator/spiff/yaml"
	"github.com/cloudfoundry-incubator/spiff/debug"
)

type MergeExpr struct {
	Path []string
	Redirect bool
}

func (e MergeExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	var info EvaluationInfo
	if e.Redirect {
		info.RedirectPath=e.Path
	}
	debug.Debug("/// lookup %v\n",e.Path)
	node, ok := binding.FindInStubs(e.Path)
	return node, info, ok
}

func (e MergeExpr) String() string {
	if e.Redirect {
		return "merge " + strings.Join(e.Path, ".")
	}
	return "merge"
}

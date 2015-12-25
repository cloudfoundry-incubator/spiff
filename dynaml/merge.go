package dynaml

import (
	"strings"
	"github.com/cloudfoundry-incubator/spiff/yaml"
	"github.com/cloudfoundry-incubator/spiff/debug"
)

type MergeExpr struct {
	Path []string
	Redirect bool
	Replace  bool
}

func (e MergeExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	var info EvaluationInfo
	if e.Redirect {
		info.RedirectPath=e.Path
	}
	debug.Debug("/// lookup %v\n",e.Path)
	node, ok := binding.FindInStubs(e.Path)
	if ok {
		info.Replace=e.Replace
	}
	return node, info, ok
}

func (e MergeExpr) String() string {
	rep := ""
	if e.Replace {
		rep = " replace"
	}
	if e.Redirect {
		return "merge" + rep + " " + strings.Join(e.Path, ".")
	}
	return "merge"+rep
}

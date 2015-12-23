package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
	"github.com/cloudfoundry-incubator/spiff/debug"
)

type MergeExpr struct {
	Path []string
}

func (e MergeExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info := EvaluationInfo{e.Path}
	debug.Debug("/// lookup %v\n",e.Path)
	node, ok := binding.FindInStubs(e.Path)
	return node, info, ok
}

func (e MergeExpr) String() string {
	return "merge"
}

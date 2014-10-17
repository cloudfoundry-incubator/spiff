package dynaml

import (
	"github.com/shutej/spiff/yaml"
)

type MergeExpr struct {
	Path []string
}

func (e MergeExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	return binding.FindInStubs(e.Path)
}

func (e MergeExpr) RequiresPhases() StringSet {
	// XXX(j): This requires thinking through.
	return StringSet(nil)
}

func (e MergeExpr) String() string {
	return "merge"
}

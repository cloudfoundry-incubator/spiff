package dynaml

import (
	"github.com/shutej/spiff/yaml"
)

type AutoExpr struct {
	Path []string
}

func (e AutoExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	if len(e.Path) == 3 && e.Path[0] == "resource_pools" && e.Path[2] == "size" {
		jobs, found := binding.FindFromRoot([]string{"jobs"})
		if !found {
			return nil, false
		}

		jobsList, ok := jobs.Value().([]yaml.Node)
		if !ok {
			return nil, false
		}

		var size int64

		for _, job := range jobsList {
			poolName, ok := yaml.FindString(job, "resource_pool")
			if !ok {
				continue
			}

			if poolName != e.Path[1] {
				continue
			}

			instances, ok := yaml.FindInt(job, "instances")
			if !ok {
				return nil, false
			}

			size += instances
		}

		return node(size), true
	}

	return nil, false
}

func (e AutoExpr) String() string {
	return "auto"
}

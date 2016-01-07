package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type AutoExpr struct {
	Path []string
}

func (e AutoExpr) Evaluate(binding Binding) (yaml.Node, EvaluationInfo, bool) {
	info := DefaultInfo()

	if len(e.Path) == 3 && e.Path[0] == "resource_pools" && e.Path[2] == "size" {
		jobs, found := binding.FindFromRoot([]string{"jobs"})
		if !found {
			return nil, info, false
		}

		if !isResolved(jobs) {
			return node(e), info, true
		}
		jobsList, ok := jobs.Value().([]yaml.Node)
		if !ok {
			return nil, info, false
		}

		var size int64

		for _, job := range jobsList {
			poolName, ok := yaml.FindString(job, "resource_pool")
			if !ok {
				continue
			}

			if poolName != yaml.PathComponent(e.Path[1]) {
				continue
			}

			instances, ok := yaml.FindInt(job, "instances")
			if !ok {
				return nil, info, false
			}

			size += instances
		}

		return node(size), info, true
	}

	return nil, info, false
}

func (e AutoExpr) String() string {
	return "auto"
}

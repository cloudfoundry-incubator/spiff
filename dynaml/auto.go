package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type AutoExpr struct {
	Path []string
}

func (e AutoExpr) Evaluate(context Context) yaml.Node {
	if len(e.Path) == 3 && e.Path[0] == "resource_pools" && e.Path[2] == "size" {

		jobs := context.FindFromRoot([]string{"jobs"})
		if jobs == nil {
			return nil
		}

		jobsList, ok := jobs.([]yaml.Node)
		if !ok {
			return nil
		}

		size := 0

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
				return nil
			}

			size += instances
		}

		return size
	}

	return nil
}

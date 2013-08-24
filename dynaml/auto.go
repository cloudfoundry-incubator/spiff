package dynaml

import (
	"github.com/vito/spiff/yaml"
)

type AutoExpr struct {
	Path []string
}

func (e AutoExpr) Evaluate(context Context) yaml.Node {
	if len(e.Path) == 3 && e.Path[0] == "resource_pools" && e.Path[2] == "size" {
		size := 0

		jobs := context.FindFromRoot([]string{"jobs"})
		if jobs == nil {
			return nil
		}

		jobsList, ok := jobs.([]yaml.Node)
		if !ok {
			return nil
		}

		for _, job := range jobsList {
			job, ok := job.(map[string]yaml.Node)
			if !ok {
				continue
			}

			resourcePool, ok := job["resource_pool"]
			if !ok {
				continue
			}

			poolName, ok := resourcePool.(string)
			if !ok {
				continue
			}

			if poolName != e.Path[1] {
				continue
			}

			instances, ok := job["instances"]
			if !ok {
				return nil
			}

			instanceCount, ok := instances.(int)
			if !ok {
				return nil
			}

			size += instanceCount
		}

		return size
	}

	return nil
}

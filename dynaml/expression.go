package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type Binding interface {
	FindFromRoot([]string) (yaml.Node, bool)
	FindReference([]string) (yaml.Node, bool)
	FindInStubs([]string) (yaml.Node, bool)
}

type EvaluationInfo struct {
	RedirectPath []string
	Replace bool
}

func DefaultInfo() EvaluationInfo {
	return EvaluationInfo{nil,false}
}

type Expression interface {
	Evaluate(Binding) (yaml.Node, EvaluationInfo, bool)
}

func (i EvaluationInfo) Join(o EvaluationInfo) EvaluationInfo {
	if o.RedirectPath !=nil {
		i.RedirectPath = o.RedirectPath
	}
	i.Replace = o.Replace // replace only by directly using the merge node
	return i
}
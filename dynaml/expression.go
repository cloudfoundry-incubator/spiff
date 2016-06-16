package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/yaml"
)

type Status interface {
	error
	Issue(string) yaml.Issue
}

type SourceProvider interface {
	SourceName() string
}

type Binding interface {
	SourceProvider
	GetLocalBinding() map[string]yaml.Node
	FindFromRoot([]string) (yaml.Node, bool)
	FindReference([]string) (yaml.Node, bool)
	FindInStubs([]string) (yaml.Node, bool)

	WithScope(step map[string]yaml.Node) Binding
	WithLocalScope(step map[string]yaml.Node) Binding
	WithPath(step string) Binding
	WithSource(source string) Binding
	RedirectOverwrite(path []string) Binding

	Path() []string
	StubPath() []string

	Flow(source yaml.Node, shouldOverride bool) (yaml.Node, Status)
}

type EvaluationInfo struct {
	RedirectPath []string
	Replace      bool
	Merged       bool
	Preferred    bool
	KeyName      string
	Source       string
	Issue        yaml.Issue
}

func (e EvaluationInfo) SourceName() string {
	return e.Source
}

func DefaultInfo() EvaluationInfo {
	return EvaluationInfo{nil, false, false, false, "", "", yaml.Issue{}}
}

type Expression interface {
	Evaluate(Binding) (interface{}, EvaluationInfo, bool)
}

func (i EvaluationInfo) Join(o EvaluationInfo) EvaluationInfo {
	if o.RedirectPath != nil {
		i.RedirectPath = o.RedirectPath
	}
	i.Replace = o.Replace // replace only by directly using the merge node
	i.Preferred = i.Preferred || o.Preferred
	i.Merged = i.Merged || o.Merged
	if o.KeyName != "" {
		i.KeyName = o.KeyName
	}
	if o.Issue.Issue != "" {
		i.Issue = o.Issue
	}
	return i
}

func ResolveExpressionOrPushEvaluation(e *Expression, resolved *bool, info *EvaluationInfo, binding Binding) (interface{}, EvaluationInfo, bool) {
	val, infoe, ok := (*e).Evaluate(binding)
	if info != nil {
		infoe = (*info).Join(infoe)
	}
	if !ok {
		return nil, infoe, false
	}

	if v, ok := val.(Expression); ok {
		*e = v
		*resolved = false
		return nil, infoe, true
	} else {
		return val, infoe, true
	}
}

func ResolveIntegerExpressionOrPushEvaluation(e *Expression, resolved *bool, info *EvaluationInfo, binding Binding) (int64, EvaluationInfo, bool) {
	value, infoe, ok := ResolveExpressionOrPushEvaluation(e, resolved, info, binding)

	if value == nil {
		return 0, infoe, ok
	}

	i, ok := value.(int64)
	if ok {
		return i, infoe, true
	} else {
		infoe.Issue = yaml.NewIssue("integer operand required")
		return 0, infoe, false
	}
}

func ResolveExpressionListOrPushEvaluation(list *[]Expression, resolved *bool, info *EvaluationInfo, binding Binding) ([]interface{}, EvaluationInfo, bool) {
	values := make([]interface{}, len(*list))
	pushed := make([]Expression, len(*list))
	infoe := EvaluationInfo{}
	ok := true

	copy(pushed, *list)

	for i, _ := range pushed {
		values[i], infoe, ok = ResolveExpressionOrPushEvaluation(&pushed[i], resolved, info, binding)
		info = &infoe
		if !ok {
			return nil, infoe, false
		}
	}
	*list = pushed
	return values, infoe, true

}

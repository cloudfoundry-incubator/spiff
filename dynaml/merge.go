package dynaml

import (
	"github.com/cloudfoundry-incubator/spiff/debug"
	"github.com/cloudfoundry-incubator/spiff/yaml"
	"strings"
)

type MergeExpr struct {
	Path     []string
	Redirect bool
	Replace  bool
	Required bool
	KeyName  string
}

func (e MergeExpr) Evaluate(binding Binding) (interface{}, EvaluationInfo, bool) {
	var info EvaluationInfo
	if e.Redirect {
		info.RedirectPath = e.Path
	}
	info.KeyName = e.KeyName
	debug.Debug("/// lookup %v\n", e.Path)
	node, ok := binding.FindInStubs(e.Path)
	if ok {
		info.Replace = e.Replace
		info.Merged = true
		info.Source = node.SourceName()
		return node.Value(), info, ok
	} else {
		info.Issue = yaml.NewIssue("'%s' not found in any stub", strings.Join(e.Path, "."))
		return nil, info, false
	}
}

func (e MergeExpr) String() string {
	rep := ""
	if e.Replace {
		rep = " replace"
	}

	if e.KeyName != "" {
		rep += " on " + e.KeyName
	}
	if e.Required && !e.Redirect && rep != "" {
		rep = " required"
	}
	if e.Redirect {
		return "merge" + rep + " " + strings.Join(e.Path, ".")
	}
	return "merge" + rep
}

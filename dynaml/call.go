package dynaml

import (
	"fmt"
	"strings"

	"github.com/shutej/spiff/yaml"
)

type CallExpr struct {
	Name      string
	Arguments []Expression
}

func (e CallExpr) RequiresPhases() StringSet {
	return StringSet(nil)
}

func (e CallExpr) Evaluate(binding Binding) (yaml.Node, bool) {
	// TODO(j): Evaluate built-in functions here.
	return nil, false
}

func (e CallExpr) String() string {
	args := make([]string, len(e.Arguments))
	for i, e := range e.Arguments {
		args[i] = fmt.Sprintf("%s", e)
	}

	return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))
}

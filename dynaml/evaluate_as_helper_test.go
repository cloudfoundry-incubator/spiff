package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func EvaluateAs(expected interface{}, binding Binding) *EvaluateAsMatcher {
	return &EvaluateAsMatcher{node(expected), binding}
}

type EvaluateAsMatcher struct {
	Expected yaml.Node
	Binding  Binding
}

func (matcher *EvaluateAsMatcher) Match(source interface{}) (success bool, message string, err error) {
	if source == nil && matcher.Expected == nil {
		return false, "", fmt.Errorf("Refusing to compare <nil> to <nil>.")
	}

	expr, ok := source.(Expression)
	if !ok {
		return false, "", fmt.Errorf("Not an expression: %v\n", source)
	}

	actual, ok := expr.Evaluate(matcher.Binding)
	if actual == nil || !ok {
		return false, "", fmt.Errorf("Node failed to evaluate.")
	}

	if actual.EquivalentToNode(matcher.Expected) {
		return true, formatMessage(actual, "not to evaluate to", matcher.Expected), nil
	} else {
		return false, formatMessage(actual, "to evaluate to", matcher.Expected), nil
	}

	return
}

func formatMessage(actual interface{}, message string, expected interface{}) string {
	return fmt.Sprintf("Expected %s %#v, got %#v", message, expected, actual)
}

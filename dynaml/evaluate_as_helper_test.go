package dynaml

import (
	"fmt"
	"reflect"

	"github.com/vito/spiff/yaml"
)

func EvaluateAs(expected yaml.Node, binding Binding) *EvaluateAsMatcher {
	return &EvaluateAsMatcher{expected, binding}
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
	if !ok {
		return false, "", fmt.Errorf("Node failed to evaluate.")
	}

	if reflect.DeepEqual(actual, matcher.Expected) {
		return true, formatMessage(actual, "not to evaluate to", matcher.Expected), nil
	} else {
		return false, formatMessage(actual, "to evaluate to", matcher.Expected), nil
	}

	return
}

func formatMessage(actual interface{}, message string, expected interface{}) string {
	return fmt.Sprintf("Expected %s %#v, got %#v", message, expected, actual)
}

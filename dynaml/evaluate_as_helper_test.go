package dynaml

import (
	"fmt"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func EvaluateAs(expected interface{}, binding Binding) *EvaluateAsMatcher {
	return &EvaluateAsMatcher{Expected: node(expected), Binding: binding}
}

type EvaluateAsMatcher struct {
	Expected yaml.Node
	Binding  Binding
	actual   yaml.Node
}

func (matcher *EvaluateAsMatcher) Match(source interface{}) (success bool, err error) {
	if source == nil && matcher.Expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.")
	}

	expr, ok := source.(Expression)
	if !ok {
		return false, fmt.Errorf("Not an expression: %v\n", source)
	}

	matcher.actual, _, ok = expr.Evaluate(matcher.Binding)
	if matcher.actual == nil || !ok {
		return false, fmt.Errorf("Node failed to evaluate.")
	}

	if matcher.actual.EquivalentToNode(matcher.Expected) {
		return true, nil
	} else {
		return false, nil
	}

	return
}

func formatMessage(actual interface{}, message string, expected interface{}) string {
	return fmt.Sprintf("Expected %s %#v, got %#v", message, expected, actual)
}

func (matcher *EvaluateAsMatcher) FailureMessage(actual interface{}) (message string) {
	return formatMessage(matcher.actual, "to evaluate to", matcher.Expected)
}

func (matcher *EvaluateAsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return formatMessage(matcher.actual, "not to evaluate to", matcher.Expected)
}

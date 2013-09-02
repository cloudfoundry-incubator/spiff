package dynaml

import (
	"fmt"
)

func FailToEvaluate(binding Binding) *FailToEvaluateMatcher {
	return &FailToEvaluateMatcher{binding}
}

type FailToEvaluateMatcher struct {
	Binding Binding
}

func (matcher *FailToEvaluateMatcher) Match(source interface{}) (success bool, message string, err error) {
	expr, ok := source.(Expression)
	if !ok {
		return false, "", fmt.Errorf("Not an expression: %v", source)
	}

	actual, ok := expr.Evaluate(matcher.Binding)
	if ok {
		return false, "", fmt.Errorf("Node evaluated to: %#v", actual)
	}

	return true, "", nil
}

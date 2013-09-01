package dynaml

import (
	"fmt"
)

func FailToEvaluate(context Context) *FailToEvaluateMatcher {
	return &FailToEvaluateMatcher{context}
}

type FailToEvaluateMatcher struct {
	Context Context
}

func (matcher *FailToEvaluateMatcher) Match(source interface{}) (success bool, message string, err error) {
	expr, ok := source.(Expression)
	if !ok {
		return false, "", fmt.Errorf("Not an expression: %v", source)
	}

	actual, ok := expr.Evaluate(matcher.Context)
	if ok {
		return false, "", fmt.Errorf("Node evaluated to: %#v", actual)
	}

	return true, "", nil
}

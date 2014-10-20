package dynaml

import (
	"fmt"
	"log"
	"reflect"
)

func DelayEvaluate(binding Binding) *DelayEvaluateMatcher {
	return &DelayEvaluateMatcher{binding}
}

type DelayEvaluateMatcher struct {
	Binding Binding
}

func (matcher *DelayEvaluateMatcher) Match(source interface{}) (success bool, err error) {
	expr, ok := source.(Expression)
	if !ok {
		return false, fmt.Errorf("Not an expression: %v", source)
	}

	actual, ok := expr.Evaluate(matcher.Binding)
	if !ok {
		return false, fmt.Errorf("Node failed to evaluate.")
	}

	log.Printf("%#v == %#v", actual.Value(), expr)
	if reflect.DeepEqual(actual.Value(), expr) {
		return true, nil
	} else {
		return false, fmt.Errorf("Nodes are not equivalent.")
	}

	return
}

func (matcher *DelayEvaluateMatcher) FailureMessage(actual interface{}) (message string) {
	return ""
}

func (matcher *DelayEvaluateMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return ""
}

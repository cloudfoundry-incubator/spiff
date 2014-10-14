package flow

import (
	"fmt"
	"reflect"

	"github.com/shutej/spiff/yaml"
)

func CascadeAs(expected yaml.Node, stubs ...yaml.Node) *CascadeAsMatcher {
	return &CascadeAsMatcher{Expected: expected, Stubs: stubs}
}

type CascadeAsMatcher struct {
	Expected yaml.Node
	Stubs    []yaml.Node
	actual   yaml.Node
}

func (matcher *CascadeAsMatcher) Match(source interface{}) (success bool, err error) {
	if source == nil && matcher.Expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.")
	}

	matcher.actual, err = Cascade(source.(yaml.Node), matcher.Stubs...)
	if err != nil {
		return false, err
	}

	if reflect.DeepEqual(matcher.actual, matcher.Expected) {
		return true, nil
	} else {
		return false, nil
	}

	return
}

func (matcher *CascadeAsMatcher) FailureMessage(actual interface{}) (message string) {
	return formatMessage(matcher.actual, "to flow as", matcher.Expected)
}

func (matcher *CascadeAsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return formatMessage(matcher.actual, "not to flow as", matcher.Expected)
}

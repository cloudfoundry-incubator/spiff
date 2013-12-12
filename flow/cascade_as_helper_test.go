package flow

import (
	"fmt"
	"reflect"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func CascadeAs(expected yaml.Node, stubs ...yaml.Node) *CascadeAsMatcher {
	return &CascadeAsMatcher{expected, stubs}
}

type CascadeAsMatcher struct {
	Expected yaml.Node
	Stubs    []yaml.Node
}

func (matcher *CascadeAsMatcher) Match(source interface{}) (success bool, message string, err error) {
	if source == nil && matcher.Expected == nil {
		return false, "", fmt.Errorf("Refusing to compare <nil> to <nil>.")
	}

	actual, err := Cascade(source, matcher.Stubs...)
	if err != nil {
		return false, "", err
	}

	if reflect.DeepEqual(actual, matcher.Expected) {
		return true, formatMessage(actual, "not to flow as", matcher.Expected), nil
	} else {
		return false, formatMessage(actual, "to flow as", matcher.Expected), nil
	}

	return
}

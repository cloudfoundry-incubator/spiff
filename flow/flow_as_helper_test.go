package flow

import (
	"fmt"
	"strings"

	"launchpad.net/goyaml"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func FlowAs(expected yaml.Node, stubs ...yaml.Node) *FlowAsMatcher {
	return &FlowAsMatcher{Expected: expected, Stubs: stubs}
}

type FlowAsMatcher struct {
	Expected yaml.Node
	Stubs    []yaml.Node
	actual   yaml.Node
}

func (matcher *FlowAsMatcher) Match(source interface{}) (success bool, err error) {
	if source == nil && matcher.Expected == nil {
		return false, fmt.Errorf("Refusing to compare <nil> to <nil>.")
	}

	actual, err := Flow(source.(yaml.Node), matcher.Stubs...)
	if err != nil {
		return false, err
	}

	if actual.EquivalentToNode(matcher.Expected) {
		return true, nil
	} else {
		return false, nil
	}

	return
}

func formatMessage(actual yaml.Node, message string, expected yaml.Node) string {
	return fmt.Sprintf("Expected%s\n%s%s", formatYAML(actual), message, formatYAML(expected))
}

func formatYAML(yaml yaml.Node) string {
	formatted, err := goyaml.Marshal(yaml)
	if err != nil {
		return fmt.Sprintf("\n\t<%T> %#v", yaml, yaml)
	}

	return fmt.Sprintf("\n\t%s", strings.Replace(string(formatted), "\n", "\n\t", -1))
}

func (matcher *FlowAsMatcher) FailureMessage(actual interface{}) (message string) {
	return formatMessage(matcher.actual, "to flow as", matcher.Expected)
}

func (matcher *FlowAsMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return formatMessage(matcher.actual, "not to flow as", matcher.Expected)
}

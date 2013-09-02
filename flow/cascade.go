package flow

import (
	"github.com/vito/spiff/yaml"
)

func Cascade(template yaml.Node, templates ...yaml.Node) (yaml.Node, error) {
	for i := len(templates) - 1; i >= 0; i-- {
		flowed, err := Flow(templates[i], templates[i+1:]...)
		if err != nil {
			return nil, err
		}

		templates[i] = flowed
	}

	return Flow(template, templates...)
}

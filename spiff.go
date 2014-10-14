package spiff

import (
	"fmt"
	"io/ioutil"

	"github.com/shutej/spiff/flow"
	"github.com/shutej/spiff/yaml"
)

func Merge(templateFilePath string, stubFilePaths []string) (yaml.Node, error) {
	templateFile, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		return nil, fmt.Errorf("error reading template: %s", err.Error())
	}

	templateYAML, err := yaml.Parse(templateFilePath, templateFile)
	if err != nil {
		return nil, fmt.Errorf("error parsing template: %s", err.Error())
	}

	stubs := []yaml.Node{}

	for _, stubFilePath := range stubFilePaths {
		stubFile, err := ioutil.ReadFile(stubFilePath)
		if err != nil {
			return nil, fmt.Errorf("error reading stub: %s", err.Error())
		}

		stubYAML, err := yaml.Parse(stubFilePath, stubFile)
		if err != nil {
			return nil, fmt.Errorf("error parsing stub: %s", err.Error())
		}

		stubs = append(stubs, stubYAML)
	}

	flowed, err := flow.Cascade(templateYAML, stubs...)
	if err != nil {
		return nil, fmt.Errorf("error generating manifest: %s", err.Error())
	}

	return flowed, nil
}

package spiff

import (
	"fmt"
	"io/ioutil"

	"github.com/shutej/spiff/yaml"
)

func ParseAll(paths []string) ([]yaml.Node, error) {
	nodes := []yaml.Node{}

	for _, path := range paths {
		file, err := ioutil.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("error reading path %q: %s", path, err.Error())
		}

		node, err := yaml.Parse(path, file)
		if err != nil {
			return nil, fmt.Errorf("error parsing stub: %s", err.Error())
		}

		nodes = append(nodes, node)
	}

	return nodes, nil
}

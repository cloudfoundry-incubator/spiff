package dynaml

import (
	"testing"

	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/vito/spiff/yaml"
)

func Test(t *testing.T) {
	RegisterFailHandler(d.Fail)
	d.RunSpecs(t, "Dynaml")
}

func parseYAML(source string) yaml.Node {
	parsed, err := yaml.Parse([]byte(source))
	if err != nil {
		panic(err)
	}

	return parsed
}

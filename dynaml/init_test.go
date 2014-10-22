package dynaml

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/shutej/spiff/yaml"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Dynaml")
}

func parseYAML(source string) yaml.Node {
	parsed, err := yaml.Parse("dynaml test", []byte(source))
	if err != nil {
		panic(err)
	}

	return parsed
}

package compare

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/cloudfoundry-incubator/spiff/yaml"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Diffing")
}

func parseYAML(source string) yaml.Node {
	return parseYAMLFrom(source, "compare test")
}

func parseYAMLFrom(source string, name string) yaml.Node {
	parsed, err := yaml.Parse(name, []byte(source))
	if err != nil {
		panic(err)
	}

	return parsed
}

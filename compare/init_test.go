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
	parsed, err := yaml.Parse([]byte(source))
	if err != nil {
		panic(err)
	}

	return parsed
}

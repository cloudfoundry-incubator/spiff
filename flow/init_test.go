package flow

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/shutej/spiff/yaml"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flowing")
}

func parseYAML(source string) yaml.Node {
	parsed, err := yaml.Parse("test", []byte(source))
	if err != nil {
		panic(err)
	}

	return parsed
}

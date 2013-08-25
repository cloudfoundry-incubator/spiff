package yaml

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "YAML parsing")
}

func parseYAML(source string) Node {
	parsed, err := Parse([]byte(source))
	if err != nil {
		panic(err)
	}

	return parsed
}

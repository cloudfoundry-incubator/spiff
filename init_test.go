package main

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"github.com/shutej/spiff/yaml"
)

var spiff string

func Test(t *testing.T) {
	BeforeSuite(func() {
		var err error
		spiff, err = gexec.Build("github.com/shutej/spiff")
		Ω(err).ShouldNot(HaveOccurred())
	})

	AfterSuite(func() {
		gexec.CleanupBuildArtifacts()
	})

	RegisterFailHandler(Fail)
	RunSpecs(t, "Executable")
}

func parseYAML(source string) yaml.Node {
	parsed, err := yaml.Parse("test", []byte(source))
	if err != nil {
		panic(err)
	}

	return parsed
}

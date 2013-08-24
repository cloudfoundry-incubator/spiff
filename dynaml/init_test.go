package dynaml

import (
	"testing"

	d "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func Test(t *testing.T) {
	RegisterFailHandler(d.Fail)
	d.RunSpecs(t, "Dynaml")
}

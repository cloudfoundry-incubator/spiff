package main

import (
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gbytes"
	. "github.com/onsi/gomega/gexec"
)

var _ = Describe("Running spiff", func() {
	Describe("merge", func() {
		var merge *Session

		Context("when given a bad file path", func() {
			BeforeEach(func() {
				var err error
				merge, err = Start(exec.Command(spiff, "merge", "foo.yml"), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
			})

			It("says file not found", func() {
				Expect(merge.Wait()).To(Exit(1))
				Expect(merge.Err).To(Say("foo.yml: no such file or directory"))
			})
		})

		Context("when given a single file", func() {
			var basicTemplate *os.File

			BeforeEach(func() {
				var err error

				basicTemplate, err = ioutil.TempFile(os.TempDir(), "basic.yml")
				Expect(err).NotTo(HaveOccurred())
				basicTemplate.Write([]byte(`
---
foo: bar
`))
				merge, err = Start(exec.Command(spiff, "merge", basicTemplate.Name()), GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				os.Remove(basicTemplate.Name())
			})

			It("resolves the template and prints it out", func() {
				Expect(merge.Wait()).To(Exit(0))
				Expect(merge.Out).To(Say(`foo: bar`))
			})
		})
	})
})

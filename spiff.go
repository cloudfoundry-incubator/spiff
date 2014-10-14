package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/codegangsta/cli"

	"github.com/cloudfoundry-incubator/candiedyaml"
	"github.com/shutej/spiff/compare"
	"github.com/shutej/spiff/flow"
	"github.com/shutej/spiff/yaml"
)

func main() {
	app := cli.NewApp()
	app.Name = "spiff"
	app.Usage = "BOSH deployment manifest toolkit"
	app.Version = "1.0.3"

	app.Commands = []cli.Command{
		{
			Name:      "merge",
			ShortName: "m",
			Usage:     "merge stub files into a manifest template",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 1 {
					cli.ShowCommandHelp(c, "merge")
					os.Exit(1)
				}

				merge(c.Args()[0], c.Args()[1:])
			},
		},
		{
			Name:      "diff",
			ShortName: "d",
			Usage:     "structurally compare two YAML files",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "separator",
					Usage: "separator to print between diffs",
				},
			},
			Action: func(c *cli.Context) {
				if len(c.Args()) < 2 {
					cli.ShowCommandHelp(c, "diff")
					os.Exit(1)
				}

				diff(c.Args()[0], c.Args()[1], c.String("separator"))
			},
		},
	}

	app.Run(os.Args)
}

func merge(templateFilePath string, stubFilePaths []string) {
	templateFile, err := ioutil.ReadFile(templateFilePath)
	if err != nil {
		log.Fatalln("error reading template:", err)
	}

	templateYAML, err := yaml.Parse(templateFilePath, templateFile)
	if err != nil {
		log.Fatalln("error parsing template:", err)
	}

	stubs := []yaml.Node{}

	for _, stubFilePath := range stubFilePaths {
		stubFile, err := ioutil.ReadFile(stubFilePath)
		if err != nil {
			log.Fatalln("error reading stub:", err)
		}

		stubYAML, err := yaml.Parse(stubFilePath, stubFile)
		if err != nil {
			log.Fatalln("error parsing stub:", err)
		}

		stubs = append(stubs, stubYAML)
	}

	flowed, err := flow.Cascade(templateYAML, stubs...)
	if err != nil {
		log.Fatalln("error generating manifest:", err)
	}

	yaml, err := candiedyaml.Marshal(flowed)
	if err != nil {
		log.Fatalln("error marshalling manifest:", err)
	}

	fmt.Println(string(yaml))
}

func diff(aFilePath, bFilePath string, separator string) {
	aFile, err := ioutil.ReadFile(aFilePath)
	if err != nil {
		log.Fatalln("error reading a:", err)
	}

	aYAML, err := yaml.Parse(aFilePath, aFile)
	if err != nil {
		log.Fatalln("error parsing a:", err)
	}

	bFile, err := ioutil.ReadFile(bFilePath)
	if err != nil {
		log.Fatalln("error reading b:", err)
	}

	bYAML, err := yaml.Parse(bFilePath, bFile)
	if err != nil {
		log.Fatalln("error parsing b:", err)
	}

	diffs := compare.Compare(aYAML, bYAML)

	if len(diffs) == 0 {
		fmt.Println("no differences!")
		return
	}

	for _, diff := range diffs {
		fmt.Println("Difference in", strings.Join(diff.Path, "."))

		if diff.A != nil {
			ayaml, err := candiedyaml.Marshal(diff.A)
			if err != nil {
				panic(err)
			}

			fmt.Printf("  %s has:\n    \x1b[31m%s\x1b[0m\n", aFilePath, strings.Replace(string(ayaml), "\n", "\n    ", -1))
		}

		if diff.B != nil {
			byaml, err := candiedyaml.Marshal(diff.B)
			if err != nil {
				panic(err)
			}

			fmt.Printf("  %s has:\n    \x1b[32m%s\x1b[0m\n", bFilePath, strings.Replace(string(byaml), "\n", "\n    ", -1))
		}

		fmt.Printf(separator)
	}
}

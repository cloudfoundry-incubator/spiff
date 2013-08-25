package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/codegangsta/cli"

	"launchpad.net/goyaml"

	"github.com/vito/spiff/flow"
	"github.com/vito/spiff/yaml"
)

func main() {
	app := cli.NewApp()
	app.Name = "spiff"
	app.Usage = "BOSH deployment manifest toolkit"

	app.Commands = []cli.Command{
		{
			Name:      "merge",
			ShortName: "m",
			Usage:     "merge merge stub files into a manifest template",
			Action: func(c *cli.Context) {
				if len(c.Args()) < 2 {
					cli.ShowCommandHelp(c, "merge")
					os.Exit(1)
				}

				merge(c.Args()[0], c.Args()[1:])
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

	templateYAML, err := yaml.Parse(templateFile)
	if err != nil {
		log.Fatalln("error parsing template:", err)
	}

	stubs := []yaml.Node{}

	for _, stubFilePath := range stubFilePaths {
		stubFile, err := ioutil.ReadFile(stubFilePath)
		if err != nil {
			log.Fatalln("error reading stub:", err)
		}

		stubYAML, err := yaml.Parse(stubFile)
		if err != nil {
			log.Fatalln("error parsing stub:", err)
		}

		stubs = append(stubs, stubYAML)
	}

	flowed, err := flow.Flow(templateYAML, stubs...)
	if err != nil {
		log.Fatalln("error generating manifest:", err)
	}

	yaml, err := goyaml.Marshal(flowed)
	if err != nil {
		log.Fatalln("error marshalling manifest:", err)
	}

	fmt.Println(string(yaml))
}

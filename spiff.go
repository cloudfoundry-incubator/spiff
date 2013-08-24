package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"

	"launchpad.net/goyaml"

	"github.com/vito/spiff/flow"
	"github.com/vito/spiff/yaml"
)

var templateFilePath = flag.String("template", "", "path to manifest template")
var stubFilePath = flag.String("stub", "", "path to stub .yml file")

func main() {
	flag.Parse()

	templateFile, err := ioutil.ReadFile(*templateFilePath)
	if err != nil {
		log.Fatalln("error reading template:", err)
	}

	stubFile, err := ioutil.ReadFile(*stubFilePath)
	if err != nil {
		log.Fatalln("error reading stub:", err)
	}

	templateYAML, err := yaml.Parse(templateFile)
	if err != nil {
		log.Fatalln("error parsing template:", err)
	}

	stubYAML, err := yaml.Parse(stubFile)
	if err != nil {
		log.Fatalln("error parsing stub:", err)
	}

	flowed, err := flow.Flow(templateYAML, stubYAML)
	if err != nil {
		log.Fatalln("error generating manifest:", err)
	}

	yaml, err := goyaml.Marshal(flowed)
	if err != nil {
		log.Fatalln("error marshalling manifest:", err)
	}

	fmt.Println(string(yaml))
}

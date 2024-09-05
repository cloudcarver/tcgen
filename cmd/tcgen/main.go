package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cloudcarver/tcgen/pkg/core"
	"gopkg.in/yaml.v3"
)

func main() {
	var toolsFilePath string
	flag.StringVar(&toolsFilePath, "path", "", "config file path")
	flag.Parse()

	if len(toolsFilePath) == 0 {
		panic("tools file path is empty")
	}

	raw, err := os.ReadFile(toolsFilePath)
	if err != nil {
		panic(err)
	}

	var data map[string]any
	if err := yaml.Unmarshal(raw, &data); err != nil {
		panic(err)
	}

	result, err := core.Generate(data)
	if err != nil {
		panic(err)
	}

	fmt.Println(result)

}

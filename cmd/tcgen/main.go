package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cloudcarver/tcgen/pkg/config"
	"github.com/cloudcarver/tcgen/pkg/core"
	"gopkg.in/yaml.v3"
)

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	var (
		version    bool
		configPath string
	)
	flag.StringVar(&configPath, "config", "", "config file path")
	flag.BoolVar(&version, "version", false, "print version")
	flag.Parse()

	if version {
		fmt.Println("v0.3.3")
		return
	}

	cfg, err := config.NewConfig("tcgen.yaml")
	must(err)

	if len(cfg.Input.Path) == 0 {
		panic("input.path is not set")
	}
	raw, err := os.ReadFile(cfg.Input.Path)
	must(err)
	var data map[string]any
	must(yaml.Unmarshal(raw, &data))

	if cfg.OpenAPI != nil {
		var raw []byte
		if len(cfg.OpenAPI.OverrideFile) != 0 {
			raw, err = os.ReadFile(cfg.OpenAPI.OverrideFile)
			must(err)
		}
		result, err := core.GenerateOpenAPISpec(raw, data, cfg.OpenAPI)
		must(err)

		must(os.WriteFile(cfg.OpenAPI.Out, []byte(result), 0644))
	}
	core.ResetGlobalTypeNameCounter()

	if cfg.GoInterpreter != nil {
		result, err := core.GenerateToolInterfaces(cfg.GoInterpreter.Package, data)
		must(err)

		must(os.WriteFile(cfg.GoInterpreter.OutPath, []byte(result), 0644))
	}
}

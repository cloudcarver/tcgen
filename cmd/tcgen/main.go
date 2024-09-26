package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/cloudcarver/edc/conf"
	"github.com/cloudcarver/tcgen/pkg/core"
	"gopkg.in/yaml.v3"
)

type Methods struct {
	Override any `yaml:"override"`
}

type Paths struct {
	Methods Methods `yaml:"methods"`
}

type OpenAPI struct {
	OverrideFile string `yaml:"overrideFile"`
	Out          string `yaml:"out"`
	PathPrefix   string `yaml:"pathPrefix"`
	Paths        Paths  `yaml:"paths"`
}

type GoInterpreter struct {
	OutPath string `yaml:"outPath"`
	Package string `yaml:"package"`
}

type Input struct {
	Path string `yaml:"path"`
}

type Config struct {
	Input         Input          `yaml:"input"`
	OpenAPI       *OpenAPI       `yaml:"openapi"`
	GoInterpreter *GoInterpreter `yaml:"goInterpreter"`
}

func NewConfig(path string) (*Config, error) {
	cfg := &Config{}
	err := conf.FetchConfig("tcgen.yaml", "TCGEN_", cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

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
		fmt.Println("v0.3.1")
		return
	}

	cfg, err := NewConfig("tcgen.yaml")
	must(err)

	if len(cfg.Input.Path) == 0 {
		panic("input.path is not set")
	}
	raw, err := os.ReadFile(cfg.Input.Path)
	must(err)
	var data map[string]any
	must(yaml.Unmarshal(raw, &data))

	fmt.Println(cfg.OpenAPI != nil, cfg.GoInterpreter != nil)
	if cfg.OpenAPI != nil {
		var raw []byte
		if len(cfg.OpenAPI.OverrideFile) != 0 {
			raw, err = os.ReadFile(cfg.OpenAPI.OverrideFile)
			must(err)
		}
		result, err := core.GenerateOpenAPISpec(raw, data, cfg.OpenAPI.PathPrefix)
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

package config

import "github.com/cloudcarver/edc/conf"

type Methods struct {
	Override any `yaml:"override"`
}

type Paths struct {
	Methods Methods `yaml:"methods"`
	Prefix  string  `yaml:"prefix"`
	Skip    bool    `yaml:"skip"`
}

type OpenAPI struct {
	OverrideFile string `yaml:"overrideFile"`
	Out          string `yaml:"out"`
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

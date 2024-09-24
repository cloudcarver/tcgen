package config

import (
	"github.com/cloudcarver/edc/conf"
)

type Methods struct {
	Override any `yaml:"override"`
}

type Paths struct {
	Methods Methods `yaml:"methods"`
}

type OpenAPI struct {
	PathPrefix string `yaml:"pathPrefix"`
	Paths      Paths  `yaml:"paths"`
}

type GoInterpreter struct {
}

type Config struct {
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

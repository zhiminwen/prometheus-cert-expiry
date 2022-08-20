package main

import (
	"os"

	"github.com/dsnet/try"
	"gopkg.in/yaml.v2"
)

type Config struct {
	Port  int
	Certs []Cert `yaml:"certs"`
}

type Cert struct {
	Name     string `yaml:"name"`
	Type     string `yaml:"type"`
	File     string `yaml:"file"`
	Address  string `yaml:"address"`
	Interval string `yaml:"interval"`
}

func ReadConfig(filename string) Config {
	content := try.E1(os.ReadFile(filename))

	var conf Config
	try.E(yaml.Unmarshal(content, &conf))
	return conf
}

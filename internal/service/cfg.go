package service

import (
	"gopkg.in/yaml.v3"
	"os"
)

type AgentConfiguration struct {
	Vcs struct {
		URI string `yaml:"uri"`
		Ref string `yaml:"ref"`
	}
}

func NewAgentConfiguration(cfgFile string) (*AgentConfiguration, error) {
	d, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	var cfg *AgentConfiguration
	err = yaml.Unmarshal(d, &cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

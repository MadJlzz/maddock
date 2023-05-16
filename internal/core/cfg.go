package core

import (
	"gopkg.in/yaml.v3"
	"os"
	"path"
	"path/filepath"
	"time"
)

const (
	agentDefaultVcsPollDelay = time.Minute
	vcsDefaultDestination    = "."
)

type AgentConfiguration struct {
	VcsPollDelay time.Duration `yaml:"vcsPollDelay"`
	Vcs          struct {
		URI         string `yaml:"uri"`
		Ref         string `yaml:"ref"`
		Destination string `yaml:"destination"`
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
	cfg.setDefaults()

	return cfg, nil
}

func (ac *AgentConfiguration) setDefaults() {
	if ac.VcsPollDelay == 0 {
		ac.VcsPollDelay = agentDefaultVcsPollDelay
	}
	if ac.Vcs.Destination == "" {
		ac.Vcs.Destination = filepath.Join(vcsDefaultDestination, path.Base(ac.Vcs.URI))
	}
}

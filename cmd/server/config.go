package main

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Target struct {
	Hostname string `yaml:"hostname"`
	Address  string `yaml:"address"`
	Manifest string `yaml:"manifest"`
}

type Config struct {
	Server struct {
		Listen string `yaml:"listen"`
	} `yaml:"server"`
	Targets []Target `yaml:"targets"`
}

// loadConfig reads and parses the server config. Manifest paths that are
// not absolute are resolved relative to the config file's directory, so
// configs remain portable regardless of the CWD where the binary runs.
func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}
	var c Config
	if err := yaml.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	configDir := filepath.Dir(path)
	for i := range c.Targets {
		if !filepath.IsAbs(c.Targets[i].Manifest) {
			c.Targets[i].Manifest = filepath.Join(configDir, c.Targets[i].Manifest)
		}
	}
	return &c, nil
}

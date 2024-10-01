package main

import (
	"fmt"
	"gopkg.in/yaml.v3"
)

type KernelParameter struct {
	Key   string
	Value string
}

func (k *KernelParameter) String() string {
	return fmt.Sprintf("%s:%s", k.Key, k.Value)
}

func (k *KernelParameter) UnmarshalYAML(value *yaml.Node) error {
	k.Key = value.Content[0].Value
	k.Value = value.Content[1].Value
	return nil
}

type Recipe struct {
	KernelParameters []KernelParameter `yaml:"kernel_parameters"`
}

package catalog

import "github.com/MadJlzz/maddock/internal/resource"

type Catalog struct {
	Name      string              `yaml:"name"`
	Resources []resource.Resource `yaml:"resources"`
}

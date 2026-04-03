package catalog

import (
	"github.com/MadJlzz/maddock/internal/resource"
	"gopkg.in/yaml.v3"
)

type rawManifest struct {
	Name      string                                 `yaml:"name"`
	Resources []map[string]map[string]map[string]any `yaml:"resources"`
}

func Parse(data []byte) (*Catalog, error) {
	var manifest rawManifest
	err := yaml.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}

	c := Catalog{
		Name:      manifest.Name,
		Resources: make([]resource.Resource, 0),
	}
	for _, item := range manifest.Resources {
		for resourceType, resources := range item {
			for resourceName, resourceAttributes := range resources {
				parsedResource, errParsing := resource.Parse(resourceType, resourceName, resourceAttributes)
				if errParsing != nil {
					return nil, errParsing
				}
				c.Resources = append(c.Resources, parsedResource)
			}
		}
	}
	return &c, nil
}

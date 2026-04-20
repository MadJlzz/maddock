package catalog

import (
	"github.com/MadJlzz/maddock/internal/resource"
	"gopkg.in/yaml.v3"
)

type rawManifest struct {
	Name      string                                 `yaml:"name"`
	Resources []map[string]map[string]map[string]any `yaml:"resources"`
}

// RawResource is a resource as declared in the manifest,
// before any type-specific validation.
type RawResource struct {
	Type       string
	Name       string
	Attributes map[string]any
}

// RawCatalog is the neutral representation of a manifest:
// resources are not yet parsed into concrete Resource implementations.
// It's used by the server binary to forward raw resources to agents
// without needing every resource type compiled in.
type RawCatalog struct {
	Name      string
	Resources []RawResource
}

// ParseRaw walks the manifest YAML and produces a RawCatalog,
// preserving the declared order of resources.
func ParseRaw(data []byte) (*RawCatalog, error) {
	var manifest rawManifest
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}

	rc := &RawCatalog{Name: manifest.Name}
	for _, item := range manifest.Resources {
		for resourceType, resources := range item {
			for resourceName, resourceAttributes := range resources {
				rc.Resources = append(rc.Resources, RawResource{
					Type:       resourceType,
					Name:       resourceName,
					Attributes: resourceAttributes,
				})
			}
		}
	}
	return rc, nil
}

func Parse(data []byte) (*Catalog, error) {
	raw, err := ParseRaw(data)
	if err != nil {
		return nil, err
	}

	c := Catalog{
		Name:      raw.Name,
		Resources: make([]resource.Resource, 0, len(raw.Resources)),
	}
	for _, rr := range raw.Resources {
		parsed, err := resource.Parse(rr.Type, rr.Name, rr.Attributes)
		if err != nil {
			return nil, err
		}
		c.Resources = append(c.Resources, parsed)
	}
	return &c, nil
}

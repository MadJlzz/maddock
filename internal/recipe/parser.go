package recipe

import (
	"fmt"
	"github.com/MadJlzz/maddock/internal/modules"
	"gopkg.in/yaml.v3"
	"io/fs"
	"path/filepath"
)

type Recipe struct {
	Modules []modules.Module
}

func DiscoverRecipes(sourcePath string) []string {
	var recipesFilepath []string
	err := filepath.WalkDir(sourcePath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", path, err)
			return err
		}
		if !d.IsDir() && filepath.Ext(d.Name()) == ".yml" {
			fmt.Printf("visited file or dir: %q\n", path)
			recipesFilepath = append(recipesFilepath, path)
		}
		return nil
	})
	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", sourcePath, err)
		return nil
	}
	return recipesFilepath
}

type tmpRecipe struct {
	KPM modules.KernelParameterModule `yaml:"kernel_parameters"`
}

func (r *Recipe) UnmarshalYAML(value *yaml.Node) error {
	var tmp tmpRecipe
	if err := value.Decode(&tmp); err != nil {
		return err
	}
	r.Modules = append(r.Modules, &tmp.KPM)
	return nil
}

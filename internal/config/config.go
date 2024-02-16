package config

import (
	"os"
	"path/filepath"
	"slices"
	"time"

	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

type ExcludeType []Exclude

var Config ExcludeType

func (e ExcludeType) IsCheckHash(hash string) bool {
	now := time.Now()
	for _, exclude := range e {
		if exclude.Hash == hash {
			if exclude.Before == "" {
				return true
			}

			before, err := time.Parse("02.01.2006", exclude.Before)
			if err != nil {
				panic(err)
			}

			if now.Before(before) {
				return false
			}
		}
	}

	return false
}

func (e ExcludeType) IsCheckNoLint(path, name string, linters []string) bool {
	dir, err := os.Getwd()
	if err != nil {
		return false
	}

	path, _ = filepath.Rel(dir, path)
	for _, exclude := range e {
		if exclude.Path == path && exclude.Name == name {
			for _, linter := range linters {
				if slices.Contains(exclude.Linters, linter) {
					return true
				}
			}
		}
	}
	return false
}

type Exclude struct {
	Jira    string   `yaml:"Jira"`
	Comment string   `yaml:"Comment"`
	Before  string   `yaml:"Before"`
	Hash    string   `yaml:"Hash"`
	Path    string   `yaml:"Path"`
	Name    string   `yaml:"Name"`
	Linters []string `yaml:"Linters"`
}

// Read load config file `.golimiter.yaml`.
func Read() {
	var config map[string][]Exclude

	gomod := ReadModFile()

	body, err := os.ReadFile(".golimiter.yaml")
	if err != nil {
		return
	}

	if err := yaml.Unmarshal(body, &config); err != nil {
		return
	}

	Config = config[gomod.Module.Mod.String()]
}

// ReadModFile return info from file go.mod.
func ReadModFile() *modfile.File {
	body, err := os.ReadFile("go.mod")
	if err != nil {
		panic(err)
	}

	gomodfile, err := modfile.Parse("go.mod", body, nil)
	if err != nil {
		panic(err)
	}

	return gomodfile
}

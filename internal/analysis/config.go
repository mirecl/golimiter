package analysis

import (
	"os"
	"path/filepath"
	"slices"
	"time"

	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

type Config struct {
	NoNoLint    ConfigNoNoLint      `yaml:"NoNoLint"`
	NoGoroutine ConfigDefaultLinter `yaml:"NoGoroutine"`
	NoLength    ConfigDefaultLinter `yaml:"NoLength"`
	NoDefer     ConfigDefaultLinter `yaml:"NoDefer"`
	NoInit      ConfigDefaultLinter `yaml:"NoInit"`
	NoGeneric   ConfigDefaultLinter `yaml:"NoGeneric"`
}

type ConfigDefaultLinter struct {
	ExcludeHashs []ExcludeHash `yaml:"ExcludeHashs"`
	ExcludeNames []ExcludeName `yaml:"ExcludeNames"`
}

type ConfigNoNoLint struct {
	ExcludeHashs []ExcludeHash         `yaml:"ExcludeHashs"`
	ExcludeNames []ExcludeNameNoNoLint `yaml:"ExcludeNames"`
}

type ExcludeHash struct {
	Hash    string    `yaml:"Hash"`
	Before  time.Time `yaml:"Before"`
	Comment string    `yaml:"Comment"`
}

type ExcludeName struct {
	Name    string    `yaml:"Name"`
	Path    string    `yaml:"Path"`
	Before  time.Time `yaml:"Before"`
	Comment string    `yaml:"Comment"`
}

type ExcludeNameNoNoLint struct {
	Position ExcludeName `yaml:"Position"`
	Linters  []string    `yaml:"Linters"`
}

func (c ConfigNoNoLint) IsVerifyHash(hash string) bool {
	return isVerifyHash(hash, c.ExcludeHashs)
}

func (c ConfigDefaultLinter) IsVerifyHash(hash string) bool {
	return isVerifyHash(hash, c.ExcludeHashs)
}

func isVerifyHash(value string, hashs []ExcludeHash) bool {
	for _, hash := range hashs {
		if hash.IsVerify(value) {
			return true
		}
	}
	return false
}

func (c ConfigDefaultLinter) IsVerifyName(path, name string) bool {
	for _, exclude := range c.ExcludeNames {
		if exclude.IsVerify(path, name) {
			return true
		}
	}
	return false
}

func (en ExcludeName) IsVerify(path, name string) bool {
	now := time.Now()

	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path, err = filepath.Rel(dir, path)
	if err != nil {
		panic(err)
	}

	if en.Path == path && en.Name == name {
		if en.Before.IsZero() {
			return true
		}

		if now.Before(en.Before) {
			return true
		}
	}

	return false
}

func (eh ExcludeHash) IsVerify(hash string) bool {
	now := time.Now()
	if eh.Hash == hash {
		if eh.Before.IsZero() {
			return true
		}

		if now.Before(eh.Before) {
			return true
		}
	}
	return false
}

func (c ConfigNoNoLint) IsVerifyName(path, name string, linters []string) bool {
	for _, exclude := range c.ExcludeNames {
		isVerifyLinter := true
		for _, linter := range linters {
			if !slices.Contains(exclude.Linters, linter) {
				isVerifyLinter = false
			}
		}

		if isVerifyLinter && exclude.Position.IsVerify(path, name) {
			return true
		}
	}
	return false
}

// Read load config file `.golimiter.yaml`.
func ReadConfig() (*Config, error) {
	var config map[string]Config

	gomod, err := ReadModFile()
	if err != nil {
		return nil, err
	}

	body, err := os.ReadFile(".golimiter.yaml")
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(body, &config); err != nil {
		return nil, err
	}

	cfg := config[gomod.Module.Mod.String()]
	return &cfg, nil
}

// ReadModFile return info from file go.mod.
func ReadModFile() (*modfile.File, error) {
	body, err := os.ReadFile("go.mod")
	if err != nil {
		return nil, err
	}

	gomodfile, err := modfile.Parse("go.mod", body, nil)
	if err != nil {
		return nil, err
	}

	return gomodfile, nil
}

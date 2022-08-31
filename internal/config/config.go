package config

import (
	"os"

	"github.com/mirecl/golimiter/pkg/exprcheck"
	"github.com/mirecl/golimiter/pkg/goroutinecheck"
	"github.com/mirecl/golimiter/pkg/initcheck"
	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

// Config is global configuration struct.
type Config struct {
	Init      initcheck.Config      `arg:"-" yaml:"init"`
	Goroutine goroutinecheck.Config `yaml:"goroutine"`
	Expr      exprcheck.Config      `yaml:"expr"`
}

// Default returns `default` value for global Config.
func Default() *Config {
	mf := ReadModFile()

	config := &Config{
		Expr: exprcheck.Config{
			Complexity: 3,
			ModFile:    mf,
		},
	}

	return config
}

// Read load config file `.golimiter.yaml`.
func Read() *Config {
	// default config
	config := Default()

	body, err := os.ReadFile(".golimiter.yaml")
	if err != nil {
		return config
	}

	if err := yaml.Unmarshal(body, config); err != nil {
		return config
	}

	return config
}

// ReadModFile return info from file go.mod.
func ReadModFile() *modfile.File {
	body, err := os.ReadFile("go.mod")
	if err != nil {
		return nil
	}

	gomodfile, err := modfile.Parse("go.mod", body, nil)
	if err != nil {
		return nil
	}

	return gomodfile
}

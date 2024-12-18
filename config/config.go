package config

import (
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"time"

	"golang.org/x/mod/modfile"
	"gopkg.in/yaml.v3"
)

type Config struct {
	NoNoLint     NoNoLint      `yaml:"NoNoLint"`
	NoGoroutine  DefaultLinter `yaml:"NoGoroutine"`
	NoLength     DefaultLinter `yaml:"NoLength"`
	NoDefer      DefaultLinter `yaml:"NoDefer"`
	NoDoc        DefaultLinter `yaml:"NoDoc"`
	NoInit       DefaultLinter `yaml:"NoInit"`
	NoGeneric    DefaultLinter `yaml:"NoGeneric"`
	NoPrefix     DefaultLinter `yaml:"NoPrefix"`
	NoUnderscore DefaultLinter `yaml:"NoUnderscore"`
	NoObject     DefaultLinter `yaml:"NoObject"`
	NoEmbedding  DefaultLinter `yaml:"NoEmbedding"`
}

type Info struct {
	Severity string `yaml:"Severity"`
	Disable  bool   `yaml:"Disable"`
	Type     string `yaml:"Type"`
}

type DefaultLinter struct {
	ExcludeHashs   []ExcludeHash `yaml:"ExcludeHashs"`
	ExcludeNames   []ExcludeName `yaml:"ExcludeNames"`
	ExcludeFiles   []string      `yaml:"ExcludeFiles"`
	ExcludeFolders []string      `yaml:"ExcludeFolders"`
	Info           `yaml:"Info"`
}

type NoNoLint struct {
	ExcludeHashs   []ExcludeHash         `yaml:"ExcludeHashs"`
	ExcludeNames   []ExcludeNameNoNoLint `yaml:"ExcludeNames"`
	ExcludeFiles   []string              `yaml:"ExcludeFiles"`
	ExcludeFolders []string              `yaml:"ExcludeFolders"`
	Info           `yaml:"Info"`
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

func (c NoNoLint) IsVerifyHash(hash string) bool {
	return isVerifyHash(hash, c.ExcludeHashs)
}

func (c DefaultLinter) IsVerifyHash(hash string) bool {
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

func (c DefaultLinter) IsVerifyName(path, name string) bool {
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

func (c NoNoLint) IsVerifyName(path, name string, linters []string) bool {
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

type Global struct {
	ExcludeFiles   []string         `yaml:"ExcludeFiles"`
	ExcludeFolders []string         `yaml:"ExcludeFolders"`
	Linters        map[string]*Info `yaml:"Linters"`
}

type Settings struct {
	Global Global            `yaml:"global"`
	Module map[string]Config `yaml:"module"`
}

// ReadFromFile load config file `.golimiter.yaml` or stdin.
func ReadFromFile(path string) (*Config, error) {
	var body []byte
	var err error

	if path == os.Stdin.Name() {
		body, err = io.ReadAll(os.Stdin)
		if err != nil {
			return &Config{}, nil
		}
	} else {
		body, err = os.ReadFile(filepath.Clean(path))
		if err != nil {
			return &Config{}, nil
		}
	}

	return ReadFromBytes(body)
}

// ReadFromBytes load config from bytes.
func ReadFromBytes(body []byte) (*Config, error) {
	var settings Settings

	gomod, err := ReadModFile()
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(body, &settings); err != nil {
		return nil, err
	}

	cfg := settings.Module[gomod.Module.Mod.String()]

	// TODO: to be use reflect - mapping config to module
	if reflect.ValueOf(cfg.NoDefer.Info).IsZero() {
		cfg.NoDefer.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoDefer")
	}
	cfg.NoDefer.ExcludeFiles = append(cfg.NoDefer.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoDefer.ExcludeFolders = append(cfg.NoDefer.ExcludeFolders, settings.Global.ExcludeFolders...)

	if reflect.ValueOf(cfg.NoGeneric.Info).IsZero() {
		cfg.NoGeneric.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoGeneric")
	}
	cfg.NoGeneric.ExcludeFiles = append(cfg.NoGeneric.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoGeneric.ExcludeFolders = append(cfg.NoGeneric.ExcludeFolders, settings.Global.ExcludeFolders...)

	if reflect.ValueOf(cfg.NoGoroutine.Info).IsZero() {
		cfg.NoGoroutine.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoGoroutine")
	}
	cfg.NoGoroutine.ExcludeFiles = append(cfg.NoGoroutine.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoGoroutine.ExcludeFolders = append(cfg.NoGoroutine.ExcludeFolders, settings.Global.ExcludeFolders...)

	if reflect.ValueOf(cfg.NoInit.Info).IsZero() {
		cfg.NoInit.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoInit")
	}
	cfg.NoInit.ExcludeFiles = append(cfg.NoInit.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoInit.ExcludeFolders = append(cfg.NoInit.ExcludeFolders, settings.Global.ExcludeFolders...)

	if reflect.ValueOf(cfg.NoLength.Info).IsZero() {
		cfg.NoLength.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoLength")
	}
	cfg.NoLength.ExcludeFiles = append(cfg.NoLength.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoLength.ExcludeFolders = append(cfg.NoLength.ExcludeFolders, settings.Global.ExcludeFolders...)

	if reflect.ValueOf(cfg.NoNoLint.Info).IsZero() {
		cfg.NoNoLint.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoNoLint")
	}
	cfg.NoNoLint.ExcludeFiles = append(cfg.NoNoLint.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoNoLint.ExcludeFolders = append(cfg.NoNoLint.ExcludeFolders, settings.Global.ExcludeFolders...)

	if reflect.ValueOf(cfg.NoPrefix.Info).IsZero() {
		cfg.NoPrefix.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoPrefix")
	}
	cfg.NoPrefix.ExcludeFiles = append(cfg.NoPrefix.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoPrefix.ExcludeFolders = append(cfg.NoPrefix.ExcludeFolders, settings.Global.ExcludeFolders...)

	if reflect.ValueOf(cfg.NoUnderscore.Info).IsZero() {
		cfg.NoUnderscore.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoUnderscore")
	}
	cfg.NoUnderscore.ExcludeFiles = append(cfg.NoUnderscore.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoUnderscore.ExcludeFolders = append(cfg.NoUnderscore.ExcludeFolders, settings.Global.ExcludeFolders...)

	if reflect.ValueOf(cfg.NoObject.Info).IsZero() {
		cfg.NoObject.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoObject")
	}
	cfg.NoObject.ExcludeFiles = append(cfg.NoObject.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoObject.ExcludeFolders = append(cfg.NoObject.ExcludeFolders, settings.Global.ExcludeFolders...)

	if reflect.ValueOf(cfg.NoDoc.Info).IsZero() {
		cfg.NoDoc.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoDoc")
	}
	cfg.NoDoc.ExcludeFiles = append(cfg.NoDoc.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoDoc.ExcludeFolders = append(cfg.NoDoc.ExcludeFolders, settings.Global.ExcludeFolders...)

	if reflect.ValueOf(cfg.NoDoc.Info).IsZero() {
		cfg.NoEmbedding.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoEmbedding")
	}
	cfg.NoEmbedding.ExcludeFiles = append(cfg.NoEmbedding.ExcludeFiles, settings.Global.ExcludeFiles...)
	cfg.NoEmbedding.ExcludeFolders = append(cfg.NoEmbedding.ExcludeFolders, settings.Global.ExcludeFolders...)

	return &cfg, nil
}

func GetGlobalConfigForLinter(global map[string]*Info, name string) Info {
	if cfg, ok := global[name]; ok {
		if cfg != nil {
			return *cfg
		}
	}
	return Info{Severity: "BLOCKER", Disable: false, Type: "BUG"}
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

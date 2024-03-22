package analysis

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
	NoNoLint    ConfigNoNoLint      `yaml:"NoNoLint"`
	NoGoroutine ConfigDefaultLinter `yaml:"NoGoroutine"`
	NoLength    ConfigDefaultLinter `yaml:"NoLength"`
	NoDefer     ConfigDefaultLinter `yaml:"NoDefer"`
	NoInit      ConfigDefaultLinter `yaml:"NoInit"`
	NoGeneric   ConfigDefaultLinter `yaml:"NoGeneric"`
	NoPrefix    ConfigDefaultLinter `yaml:"NoPrefix"`
}

type Info struct {
	Severity string `yaml:"Severity"`
	Disable  bool   `yaml:"Disable"`
	Type     string `yaml:"Type"`
}

type ConfigDefaultLinter struct {
	ExcludeHashs []ExcludeHash `yaml:"ExcludeHashs"`
	ExcludeNames []ExcludeName `yaml:"ExcludeNames"`
	ExcludeFiles []string      `yaml:"ExcludeFiles"`
	Info         `yaml:"Info"`
}

type ConfigNoNoLint struct {
	ExcludeHashs []ExcludeHash         `yaml:"ExcludeHashs"`
	ExcludeNames []ExcludeNameNoNoLint `yaml:"ExcludeNames"`
	ExcludeFiles []string              `yaml:"ExcludeFiles"`
	Info         `yaml:"Info"`
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

type Global struct {
	ExcludeFiles []string         `yaml:"ExcludeFiles"`
	Linters      map[string]*Info `yaml:"Linters"`
}

type Settings struct {
	Global Global            `yaml:"global"`
	Module map[string]Config `yaml:"module"`
}

// ReadConfig load config file `.golimiter.yaml` or stdin.
func ReadConfig(path string) (*Config, error) {
	var settings Settings
	var body []byte

	gomod, err := ReadModFile()
	if err != nil {
		return nil, err
	}

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

	if err = yaml.Unmarshal(body, &settings); err != nil {
		return nil, err
	}

	cfg := settings.Module[gomod.Module.Mod.String()]

	// TODO: to be use reflect - mapping config to module
	if reflect.ValueOf(cfg.NoDefer.Info).IsZero() {
		cfg.NoDefer.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoDefer")
	}
	cfg.NoDefer.ExcludeFiles = append(cfg.NoDefer.ExcludeFiles, settings.Global.ExcludeFiles...)

	if reflect.ValueOf(cfg.NoGeneric.Info).IsZero() {
		cfg.NoGeneric.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoGeneric")
	}
	cfg.NoGeneric.ExcludeFiles = append(cfg.NoGeneric.ExcludeFiles, settings.Global.ExcludeFiles...)

	if reflect.ValueOf(cfg.NoGoroutine.Info).IsZero() {
		cfg.NoGoroutine.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoGoroutine")
	}
	cfg.NoGoroutine.ExcludeFiles = append(cfg.NoGoroutine.ExcludeFiles, settings.Global.ExcludeFiles...)

	if reflect.ValueOf(cfg.NoInit.Info).IsZero() {
		cfg.NoInit.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoInit")
	}
	cfg.NoInit.ExcludeFiles = append(cfg.NoInit.ExcludeFiles, settings.Global.ExcludeFiles...)

	if reflect.ValueOf(cfg.NoLength.Info).IsZero() {
		cfg.NoLength.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoLength")
	}
	cfg.NoLength.ExcludeFiles = append(cfg.NoLength.ExcludeFiles, settings.Global.ExcludeFiles...)

	if reflect.ValueOf(cfg.NoNoLint.Info).IsZero() {
		cfg.NoNoLint.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoNoLint")
	}
	cfg.NoNoLint.ExcludeFiles = append(cfg.NoNoLint.ExcludeFiles, settings.Global.ExcludeFiles...)

	if reflect.ValueOf(cfg.NoPrefix.Info).IsZero() {
		cfg.NoPrefix.Info = GetGlobalConfigForLinter(settings.Global.Linters, "NoPrefix")
	}
	cfg.NoPrefix.ExcludeFiles = append(cfg.NoPrefix.ExcludeFiles, settings.Global.ExcludeFiles...)

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

package analysis

import (
	"log"

	"golang.org/x/tools/go/packages"
)

const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedTypes |
	packages.NeedTypesSizes |
	packages.NeedTypesInfo |
	packages.NeedSyntax |
	packages.NeedModule |
	packages.NeedEmbedFiles |
	packages.NeedEmbedPatterns

// Linter describes an analysis function and its options.
type Linter struct {
	// The Name of the analyzer must be a valid Go identifier
	Name string
	// Run applies the analyzer to a package.
	Run func(*Config, []*packages.Package) []Issue
}

// Run analyze source code.
func Run(cfg *Config, linters ...*Linter) map[string][]Issue {
	pkgs, err := packages.Load(&packages.Config{Mode: loadMode, Tests: false}, "./...")
	if err != nil {
		log.Fatalf("failed load go/packages: %s", err)
	}

	allIssues := make(map[string][]Issue, len(linters))

	for _, linter := range linters {
		allIssues[linter.Name] = linter.Run(cfg, pkgs)
	}

	return allIssues
}

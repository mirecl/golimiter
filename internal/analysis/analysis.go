package analysis

import (
	"encoding/json"
	"flag"
	"fmt"
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
	Run func([]*packages.Package) []Issue
}

func Run(linters ...*Linter) {
	pkgs, err := packages.Load(&packages.Config{Mode: loadMode, Tests: false}, "./...")
	if err != nil {
		log.Fatalf("failed load go/packages: %s", err)
	}

	allIssues := make(map[string][]Issue)

	for _, linter := range linters {
		issues := linter.Run(pkgs)
		allIssues[linter.Name] = issues
	}

	jsonFlag := flag.Bool("json", false, "format report")
	flag.Parse()

	if !*jsonFlag {
		if len(allIssues) == 0 {
			fmt.Println("")
		}

		for linter, issues := range allIssues {
			if len(issues) == 0 {
				continue
			}

			for _, issue := range issues {
				fmt.Printf("%s \033[31m%s: %s (%s)\033[0m\033[30m(GOLIMITER:%s)\033[0m\n", issue.Position(), linter, issue.Message, issue.Code, issue.Hash())
			}
		}
		return
	}

	if allIssuesBytes, err := json.Marshal(allIssues); err == nil {
		fmt.Println(string(allIssuesBytes))
	}
}

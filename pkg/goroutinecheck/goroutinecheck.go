package goroutinecheck

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"github.com/mirecl/golimiter/internal"
	"github.com/mirecl/golimiter/internal/store"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

type excludeType = internal.ExcludeType

// Config linter.
type Config struct {
	ModFile      *modfile.File
	Limit        *int        `yaml:"limit"`
	ExcludeFiles excludeType `yaml:"exclude_files"`
	ExcludeFuncs excludeType `yaml:"exclude_funcs"`
}

// global state issues.
var state store.Store

// New instance linter.
func New(c *Config) *analysis.Analyzer {
	state = store.New()

	return &analysis.Analyzer{
		Name:     "goroutinecheck",
		Doc:      "Check count `goroutine` statement.",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return run(c, pass)
		},
	}
}

func run(c *Config, pass *analysis.Pass) (interface{}, error) {
	// no restrictions.
	if c.Limit == nil {
		return nil, nil
	}

	folder, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed get root path: %w", err)
	}

	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.GoStmt)(nil)}

	var pkgIssues []*store.Issue

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		var fileName, fileMod, funcName, funcMod string

		pos := pass.Fset.Position(node.Pos())

		fileName = pos.Filename
		funcName = pass.Pkg.Name() + "." + internal.GetFuncDecl(pos).Name.Name

		if c.ModFile != nil {
			fileMod = strings.ReplaceAll(fileName, folder, c.ModFile.Module.Mod.Path)
			funcMod = strings.ReplaceAll(filepath.Dir(fileName), folder, c.ModFile.Module.Mod.Path) + "/" + funcName
		}

		// check `*_test` files.
		if internal.IsTestFile(fileName) {
			return
		}

		// check exclude files.
		if c.ExcludeFiles.ConsistOf(fileName) || c.ExcludeFiles.ConsistOf(fileMod) {
			return
		}

		// check exclude funcs.
		if c.ExcludeFuncs.ConsistOf(funcName) || c.ExcludeFuncs.ConsistOf(funcMod) {
			return
		}

		pkgIssues = append(pkgIssues, &store.Issue{
			Pos:  node.Pos(),
			Pass: pass,
		})
	})

	// forbidden all `goroutine` statement.
	if *c.Limit == 0 {
		for _, issue := range pkgIssues {
			issue.Report("a `goroutine` statement forbidden to use.")
		}
		return nil, nil
	}

	for _, issue := range pkgIssues {
		state.Add(issue)
	}

	// limit `goroutine` statement.
	if state.Len() >= *c.Limit {
		state.Reportf("a number of allowed `goroutine` statement %d.", *c.Limit)
		return nil, nil
	}

	return nil, nil
}

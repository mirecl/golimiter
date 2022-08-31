package exprcheck

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"strings"

	"github.com/mirecl/golimiter/internal"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

type excludeType = internal.ExcludeType

// Config linter.
type Config struct {
	ModFile      *modfile.File
	ExcludeFiles excludeType `yaml:"exclude_files"`
	ExcludeFuncs excludeType `yaml:"exclude_funcs"`
	Complexity   int         `yaml:"complexity"`
}

// New instance linter.
func New(c *Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "exprcheck",
		Doc:      "Check complexity `expr` statement.",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return run(pass, c)
		},
	}
}

func run(pass *analysis.Pass, c *Config) (interface{}, error) {
	folder, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed get root path: %w", err)
	}

	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.IfStmt)(nil),
		(*ast.ReturnStmt)(nil),
	}

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

		r := result{}

		switch node := node.(type) {
		case *ast.IfStmt:
			r.check(node.Cond)
		case *ast.ReturnStmt:
			for _, stmt := range node.Results {
				r.check(stmt)
			}
		}

		if r.Complexity > c.Complexity {
			pass.Reportf(node.Pos(), "a complexity expr %d, allowed %d.", r.Complexity, c.Complexity)
		}
	})

	return nil, nil
}

type result struct {
	Complexity int
}

func (r *result) check(expr ast.Expr) {
	switch n := expr.(type) {
	case *ast.BinaryExpr:
		r.Complexity++
		r.check(n.X)
		r.check(n.Y)
	case *ast.ParenExpr:
		r.check(n.X)
	}
}

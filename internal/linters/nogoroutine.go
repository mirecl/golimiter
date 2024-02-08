package linters

import (
	"go/ast"
	"go/token"
	"go/types"

	"github.com/mirecl/golimiter/internal/analysis"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoGoroutine = "a `goroutine` statement forbidden to use"
	codeNoNoGoroutine  = "GL0003"
)

// NewNoGoroutine create instance linter for check goroutines.
func NewNoGoroutine() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoGoroutine",
		Run: func(pkgs []*packages.Package) []Issue {
			issues := make([]Issue, 0)

			for _, p := range pkgs {
				pkgIssues := runNoGoroutine(p.Syntax, p.TypesInfo, p.Fset)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoGoroutine(pkgFiles []*ast.File, _ *types.Info, fset *token.FileSet) []Issue {
	nodeFilter := []ast.Node{(*ast.GoStmt)(nil)}

	inspect := inspector.New(pkgFiles)

	var pkgIssues []Issue

	ignoreObjects := GetIgnore(pkgFiles, fset)

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		if ignoreObjects.IsCheck(node) {
			return
		}

		position := fset.Position(node.Pos())

		pkgIssues = append(pkgIssues, Issue{
			Pos:      node.Pos(),
			End:      node.End(),
			Fset:     fset,
			Message:  messageNoGoroutine,
			Code:     codeNoNoGoroutine,
			Line:     position.Line,
			Filename: position.Filename,
		})
	})

	return pkgIssues
}

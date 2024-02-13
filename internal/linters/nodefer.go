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
	messageNoDefer = "a `defer` statement forbidden to use"
)

// NewNoDefer create instance linter for check defer.
func NewNoDefer() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoDefer",
		Run: func(pkgs []*packages.Package) []Issue {
			issues := make([]Issue, 0)

			for _, p := range pkgs {
				pkgIssues := runNoDefer(p.Syntax, p.TypesInfo, p.Fset)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoDefer(pkgFiles []*ast.File, _ *types.Info, fset *token.FileSet) []Issue {
	nodeFilter := []ast.Node{(*ast.DeferStmt)(nil)}

	inspect := inspector.New(pkgFiles)

	var pkgIssues []Issue

	ignoreObjects := GetIgnore(pkgFiles, fset)

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		if ignoreObjects.IsCheck(node) {
			return
		}

		position := fset.Position(node.Pos())

		pkgIssues = append(pkgIssues, Issue{
			Message:  messageNoGoroutine,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     analysis.GetHash(fset, node.Pos(), node.End()),
		})
	})

	return pkgIssues
}

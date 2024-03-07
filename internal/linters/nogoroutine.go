package linters

import (
	"go/ast"

	"github.com/mirecl/golimiter/internal/analysis"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoGoroutine = "a `goroutine` statement forbidden to use"
)

// NewNoGoroutine create instance linter for check goroutines.
func NewNoGoroutine() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoGoroutine",
		Run: func(cfg *analysis.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			for _, pkg := range pkgs {
				pkgIssues := runNoGoroutine(&cfg.NoGoroutine, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

// TODO: check goroutine in func with name.
func runNoGoroutine(cfg *analysis.ConfigDefaultLinter, pkg *packages.Package) []analysis.Issue {
	nodeFilter := []ast.Node{(*ast.GoStmt)(nil)}

	inspect := inspector.New(pkg.Syntax)

	var pkgIssues []analysis.Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		hash := analysis.GetHashFromBody(pkg.Fset, node)
		if cfg.IsVerifyHash(hash) {
			return
		}

		position := pkg.Fset.Position(node.Pos())

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  messageNoGoroutine,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     hash,
			Severity: cfg.Severity,
		})
	})

	return pkgIssues
}

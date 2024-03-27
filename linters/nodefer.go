package linters

import (
	"go/ast"
	"slices"

	"github.com/mirecl/golimiter/analysis"
	"github.com/mirecl/golimiter/config"
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
		Run: func(cfg *config.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoDefer.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				pkgIssues := runNoDefer(&cfg.NoDefer, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

// TODO: check defer in func with name.
func runNoDefer(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	nodeFilter := []ast.Node{(*ast.DeferStmt)(nil)}

	inspect := inspector.New(pkg.Syntax)

	var pkgIssues []analysis.Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		position := pkg.Fset.Position(node.Pos())

		currentFile := analysis.GetPathRelative(position.Filename)
		if slices.Contains(cfg.ExcludeFiles, currentFile) {
			return
		}

		hash := analysis.GetHashFromBody(pkg.Fset, node)
		if cfg.IsVerifyHash(hash) {
			return
		}

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  messageNoDefer,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     hash,
			Severity: cfg.Severity,
			Type:     cfg.Type,
			Rule:     "nodefer",
		})
	})

	return pkgIssues
}

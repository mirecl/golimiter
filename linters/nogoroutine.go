package linters

import (
	"go/ast"
	"slices"
	"strings"

	"github.com/mirecl/golimiter/analysis"
	"github.com/mirecl/golimiter/config"
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
		Run: func(cfg *config.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoGoroutine.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				pkgIssues := runNoGoroutine(&cfg.NoGoroutine, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

// TODO: check goroutine in func with name.
func runNoGoroutine(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	nodeFilter := []ast.Node{(*ast.GoStmt)(nil)}

	inspect := inspector.New(pkg.Syntax)

	var pkgIssues []analysis.Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		position := pkg.Fset.Position(node.Pos())

		currentFile := analysis.GetPathRelative(position.Filename)
		if slices.Contains(cfg.ExcludeFiles, currentFile) {
			return
		}

		for _, folder := range cfg.ExcludeFolders {
			if strings.HasPrefix(currentFile, folder) {
				return
			}
		}

		hash := analysis.GetHashFromBody(pkg.Fset, node)
		if cfg.IsVerifyHash(hash) {
			return
		}

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  messageNoGoroutine,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     hash,
			Severity: cfg.Severity,
			Type:     cfg.Type,
		})
	})

	return pkgIssues
}

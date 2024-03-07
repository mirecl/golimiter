package linters

import (
	"go/ast"

	"github.com/mirecl/golimiter/internal/analysis"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoInit = "a `init` funcs forbidden to use"
)

// NewNoInit create instance linter for check func init.
//
//nolint:dupl
func NewNoInit() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoInit",
		Run: func(cfg *analysis.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoInit.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				pkgIssues := runNoInit(&cfg.NoInit, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoInit(cfg *analysis.ConfigDefaultLinter, pkg *packages.Package) []analysis.Issue {
	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	inspect := inspector.New(pkg.Syntax)

	var pkgIssues []analysis.Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		fn, _ := node.(*ast.FuncDecl)

		if fn.Name != nil && fn.Name.String() != "init" {
			return
		}

		hash := analysis.GetHashFromBody(pkg.Fset, node)
		if cfg.IsVerifyHash(hash) {
			return
		}

		position := pkg.Fset.Position(node.Pos())

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  messageNoInit,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     hash,
			Severity: cfg.Severity,
			Type:     cfg.Type,
		})
	})

	if len(pkgIssues) == 1 {
		return []analysis.Issue{}
	}

	return pkgIssues
}

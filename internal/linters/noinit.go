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
	messageNoInit = "a `init` funcs forbidden to use"
)

// NewNoInit create instance linter for check func init.
func NewNoInit() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoInit",
		Run: func(pkgs []*packages.Package) []Issue {
			issues := make([]Issue, 0)

			for _, p := range pkgs {
				pkgIssues := runNoInit(p.Syntax, p.TypesInfo, p.Fset)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoInit(pkgFiles []*ast.File, _ *types.Info, fset *token.FileSet) []Issue {
	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	inspect := inspector.New(pkgFiles)

	var pkgIssues []Issue

	ignoreObjects := GetIgnore(pkgFiles, fset)

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		fn, _ := node.(*ast.FuncDecl)

		if fn.Name != nil && fn.Name.String() != "init" {
			return
		}

		if ignoreObjects.IsCheck(node) {
			return
		}

		position := fset.Position(node.Pos())

		pkgIssues = append(pkgIssues, Issue{
			Message:  messageNoInit,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     analysis.GetHash(fset, node.Pos(), node.End()),
		})
	})

	if len(pkgIssues) == 1 {
		return []Issue{}
	}

	return pkgIssues
}

package linters

import (
	"go/ast"
	"go/types"

	"github.com/mirecl/golimiter/internal/analysis"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoGeneric = "a `generic` statement forbidden to use"
)

// NewNoGeneric create instance linter for check generic.
// Please more info in https://cs.opensource.google/go/x/tools/+/master:go/analysis/passes/usesgenerics/
func NewNoGeneric() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoGeneric",
		Run: func(cfg *analysis.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			for _, pkg := range pkgs {
				pkgIssues := runNoGeneric(&cfg.NoGeneric, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoGeneric(cfg *analysis.ConfigDefaultLinter, pkg *packages.Package) []analysis.Issue {
	nodeFilter := []ast.Node{
		(*ast.FuncType)(nil),
		(*ast.InterfaceType)(nil),
		(*ast.TypeSpec)(nil),
	}

	inspect := inspector.New(pkg.Syntax)

	var pkgIssues []analysis.Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		if !IsGeneric(node, pkg.TypesInfo) {
			return
		}

		hash := analysis.GetHashFromBody(pkg.Fset, node)
		if cfg.IsVerifyHash(hash) {
			return
		}

		position := pkg.Fset.Position(node.Pos())

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  messageNoGeneric,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     hash,
		})
	})

	return pkgIssues
}

// IsGeneric check code on generic.
func IsGeneric(node ast.Node, info *types.Info) bool {
	var isGeneric bool

	switch n := node.(type) {
	case *ast.FuncType:
		if tparams := ForFuncType(n); tparams != nil {
			isGeneric = true
		}
	case *ast.InterfaceType:
		tv := info.Types[n]
		if iface, _ := tv.Type.(*types.Interface); iface != nil && !iface.IsMethodSet() {
			isGeneric = true
		}
	case *ast.TypeSpec:
		if tparams := ForTypeSpec(n); tparams != nil {
			isGeneric = true
		}
	default:
		isGeneric = false
	}

	return isGeneric
}

// ForTypeSpec returns TypeParams.
func ForTypeSpec(n *ast.TypeSpec) *ast.FieldList {
	if n == nil {
		return nil
	}
	return n.TypeParams
}

// ForFuncType returns TypeParams.
func ForFuncType(n *ast.FuncType) *ast.FieldList {
	if n == nil {
		return nil
	}
	return n.TypeParams
}

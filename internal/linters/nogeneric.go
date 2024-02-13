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
	messageNoGeneric = "a `generic` statement forbidden to use"
)

// NewNoGeneric create instance linter for check generic.
// Please more info in https://cs.opensource.google/go/x/tools/+/master:go/analysis/passes/usesgenerics/
func NewNoGeneric() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoGeneric",
		Run: func(pkgs []*packages.Package) []Issue {
			issues := make([]analysis.Issue, 0)

			for _, p := range pkgs {
				pkgIssues := runNoGeneric(p.Syntax, p.TypesInfo, p.Fset)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoGeneric(pkgFiles []*ast.File, info *types.Info, fset *token.FileSet) []Issue {
	nodeFilter := []ast.Node{
		(*ast.FuncType)(nil),
		(*ast.InterfaceType)(nil),
		(*ast.TypeSpec)(nil),
	}

	ignoreObjects := GetIgnore(pkgFiles, fset)

	inspect := inspector.New(pkgFiles)

	var pkgIssues []Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		if !IsGeneric(node, info) {
			return
		}

		if ignoreObjects.IsCheck(node) {
			return
		}

		position := fset.Position(node.Pos())

		pkgIssues = append(pkgIssues, Issue{
			Message:  messageNoGeneric,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     analysis.GetHash(fset, node.Pos(), node.End()),
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

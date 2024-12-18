package linters

import (
	"go/ast"
	"go/types"
	"slices"
	"strings"

	"github.com/mirecl/golimiter/analysis"
	"github.com/mirecl/golimiter/config"
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
		Run: func(cfg *config.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoGeneric.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				pkgIssues := runNoGeneric(&cfg.NoGeneric, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoGeneric(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {

	nodeFilter := []ast.Node{
		(*ast.FuncType)(nil),
		(*ast.InterfaceType)(nil),
		(*ast.TypeSpec)(nil),
		(*ast.Ident)(nil),
	}

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

		if !IsGeneric(node, pkg.TypesInfo) {
			return
		}

		hash := analysis.GetHashFromBody(pkg.Fset, node)
		if cfg.IsVerifyHash(hash) {
			return
		}

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  messageNoGeneric,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     hash,
			Severity: cfg.Severity,
			Type:     cfg.Type,
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

		if n.Methods != nil && len(n.Methods.List) == 0 {
			isGeneric = true
		}

		iface, _ := tv.Type.(*types.Interface)

		if iface != nil && !iface.IsMethodSet() {
			isGeneric = true
		}
	case *ast.TypeSpec:
		if tparams := ForTypeSpec(n); tparams != nil {
			isGeneric = true
		}
	case *ast.Ident:
		if n.Name == "any" {
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

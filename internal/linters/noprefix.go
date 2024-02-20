package linters

import (
	"fmt"
	"go/ast"
	"strings"

	"github.com/mirecl/golimiter/internal/analysis"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoPrefix       = "a `prefix` funcs forbidden to use"
	messageNoPrefixLambda = "a `lambda` funcs forbidden to use"
)

var PrefixAllow = []string{"get", "new", "is", "calc", "validate", "normalize",
	"execute", "get", "set", "parse", "apply", "append", "clear", "remove",
	"delete", "update", "to", "from", "run", "read"}

// NewNoInit create instance linter for check func init.
//
//nolint:dupl
func NewNoPrefix() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoPrefix",
		Run: func(cfg *analysis.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			for _, pkg := range pkgs {
				pkgIssues := runNoPrefix(&cfg.NoPrefix, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoPrefix(cfg *analysis.ConfigDefaultLinter, pkg *packages.Package) []analysis.Issue {
	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	inspect := inspector.New(pkg.Syntax)

	var pkgIssues []analysis.Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		fn, _ := node.(*ast.FuncDecl)
		position := pkg.Fset.Position(node.Pos())

		if fn.Name == nil {
			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  messageNoPrefixLambda,
				Line:     position.Line,
				Filename: position.Filename,
				Hash:     "",
			})
			return
		}

		name := strings.ToLower(fn.Name.Name)

		if name == "main" {
			return
		}

		for _, prefix := range PrefixAllow {
			if strings.HasPrefix(name, prefix) {
				return
			}
		}

		filename := analysis.GetPathRelative(position.Filename)
		fmt.Println(fmt.Sprintf("%s:%d", filename, position.Line), fn.Name.Name)

		hash := analysis.GetHashFromString(fn.Name.Name)
		if cfg.IsVerifyHash(hash) {
			return
		}

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  messageNoPrefix,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     hash,
		})
	})

	return pkgIssues
}

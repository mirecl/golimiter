package linters

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/mirecl/golimiter/analysis"
	"github.com/mirecl/golimiter/config"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoUnderscorePackages = "please not use symbol `_` in package name `%s` (https://go.dev/blog/package-names)"
	messageNoUnderscoreVariable = "please not use symbol `_` in variable `%s`"
	messageNoUnderscoreType     = "please not use symbol `_` in type `%s`"
)

func NewNoUnderscore() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoUnderscore",
		Run: func(cfg *config.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoUnderscore.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				pkgIssues := runNoUnderscore(&cfg.NoUnderscore, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoUnderscore(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	var pkgIssues []analysis.Issue

	for _, s := range pkg.Syntax {
		for _, object := range s.Scope.Objects {
			if object.Name == "_" {
				continue
			}

			if !strings.Contains(object.Name, "_") {
				continue
			}

			message := ""
			switch object.Kind {
			case ast.Var:
				message = messageNoUnderscoreVariable
			case ast.Typ:
				message = messageNoUnderscoreType
			default:
				continue
			}

			position := pkg.Fset.Position(object.Pos())
			hash := analysis.GetHashFromString(object.Name)
			if cfg.IsVerifyHash(hash) {
				return pkgIssues
			}

			currentFile := analysis.GetPathRelative(position.Filename)
			if slices.Contains(cfg.ExcludeFiles, currentFile) {
				continue
			}

			isFind := false
		L:
			for _, folder := range cfg.ExcludeFolders {
				if strings.HasPrefix(currentFile, folder) {
					isFind = true
					break L
				}
			}

			if isFind {
				continue
			}

			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  fmt.Sprintf(message, object.Name),
				Line:     position.Line,
				Filename: position.Filename,
				Hash:     hash,
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
		}
	}

	nodeFilter := []ast.Node{(*ast.Ident)(nil)}
	inspect := inspector.New(pkg.Syntax)

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		position := pkg.Fset.Position(node.Pos())
		ident := node.(*ast.Ident)

		if ident.Obj == nil {
			return
		}

		if ident.Obj.Kind != ast.Var {
			return
		}

		if ident.Obj.Name == "_" {
			return
		}

		if !strings.Contains(ident.Obj.Name, "_") {
			return
		}

		currentFile := analysis.GetPathRelative(position.Filename)
		if slices.Contains(cfg.ExcludeFiles, currentFile) {
			return
		}

		for _, folder := range cfg.ExcludeFolders {
			if strings.HasPrefix(currentFile, folder) {
				return
			}
		}

		hash := analysis.GetHashFromString(ident.Obj.Name)
		if cfg.IsVerifyHash(hash) {
			return
		}

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  fmt.Sprintf(messageNoUnderscoreVariable, ident.Obj.Name),
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     hash,
			Severity: cfg.Severity,
			Type:     cfg.Type,
		})
	})

	if !strings.Contains(pkg.Name, "_") {
		return pkgIssues
	}

	hash := analysis.GetHashFromString(pkg.Name)
	if cfg.IsVerifyHash(hash) {
		return pkgIssues
	}

	for _, filename := range pkg.GoFiles {
		currentFile := analysis.GetPathRelative(filename)
		if slices.Contains(cfg.ExcludeFiles, currentFile) {
			continue
		}

		isFind := false
	K:
		for _, folder := range cfg.ExcludeFolders {
			if strings.HasPrefix(currentFile, folder) {
				isFind = true
				break K
			}
		}

		if isFind {
			continue
		}

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  fmt.Sprintf(messageNoUnderscorePackages, pkg.Name),
			Line:     1,
			Filename: filename,
			Hash:     hash,
			Severity: cfg.Severity,
			Type:     cfg.Type,
		})
	}

	return pkgIssues
}

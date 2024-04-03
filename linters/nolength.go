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
	MaxLengthObject = 30
	MaxSegmentCount = 6
)

const (
	messageNoLengthLength  = "Maximum allowed length of identifier is"
	messageNoLengthSegment = "Maximum allowed number of segments in identifier is"
)

// NewNoLength create instance linter for length object.
func NewNoLength() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoLength",
		Run: func(cfg *config.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoLength.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				pkgIssues := runNoLength(&cfg.NoLength, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

// TODO: add support ignore hash.
func runNoLength(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
		(*ast.Field)(nil),
		(*ast.FuncDecl)(nil),
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

		name := GetObjectName(node)
		if name == "" {
			return
		}

		hash := analysis.GetHashFromString(name)
		if cfg.IsVerifyHash(hash) {
			return
		}

		if len(name) > MaxLengthObject {
			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  fmt.Sprintf("%s %d (now %d)", messageNoLengthLength, MaxLengthObject, len(name)),
				Line:     position.Line,
				Filename: position.Filename,
				Hash:     hash,
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
		}

		segment := GetSegmentCount(name)
		if segment > MaxSegmentCount {
			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  fmt.Sprintf("%s %d (now %d)", messageNoLengthSegment, MaxSegmentCount, segment),
				Line:     position.Line,
				Filename: position.Filename,
				Hash:     analysis.GetHashFromString(name),
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
		}
	})
	return pkgIssues
}

func GetObjectName(node ast.Node) string {
	switch n := node.(type) {
	case *ast.TypeSpec:
		if n.Name == nil {
			return ""
		}
		return n.Name.Name
	case *ast.Field:
		if len(n.Names) == 0 {
			return ""
		}
		return n.Names[0].Name
	case *ast.FuncDecl:
		if n.Name == nil {
			return ""
		}
		return n.Name.Name
	}
	return ""
}

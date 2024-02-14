package linters

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"unicode"

	"github.com/mirecl/golimiter/internal/analysis"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	MaxLengthObject = 30
	MaxSegmentCount = 6
)

const (
	messageNoLength  = "Maximum allowed length of identifier is"
	messageNoSegment = "Maximum allowed number of segments in identifier is"
)

// NewNoLength create instance linter for length object.
func NewNoLength() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoLength",
		Run: func(pkgs []*packages.Package) []Issue {
			issues := make([]Issue, 0)

			for _, p := range pkgs {
				pkgIssues := runNoLength(p.Syntax, p.TypesInfo, p.Fset)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

// TODO: add support ignore hash
func runNoLength(pkgFiles []*ast.File, _ *types.Info, fset *token.FileSet) []Issue {
	nodeFilter := []ast.Node{
		(*ast.TypeSpec)(nil),
		(*ast.Field)(nil),
		(*ast.Ident)(nil),
	}

	inspect := inspector.New(pkgFiles)

	var pkgIssues []Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		name := GetObjectName(node)
		if name == "" {
			return
		}

		if len(name) > MaxLengthObject {
			position := fset.Position(node.Pos())

			pkgIssues = append(pkgIssues, Issue{
				Message:  fmt.Sprintf("%s %d (now %d)", messageNoLength, MaxLengthObject, len(name)),
				Line:     position.Line,
				Filename: position.Filename,
				Hash:     analysis.GetHashFromString(name),
			})
		}

		segment := GetSegmentCount(name)
		if segment > MaxSegmentCount {
			position := fset.Position(node.Pos())

			pkgIssues = append(pkgIssues, Issue{
				Message:  fmt.Sprintf("%s %d (now %d)", messageNoSegment, MaxSegmentCount, segment),
				Line:     position.Line,
				Filename: position.Filename,
				Hash:     analysis.GetHashFromString(name),
			})
		}
	})
	return pkgIssues
}

func GetSegmentCount(text string) int {
	c := 0
	isLastSymbolUpper := true
	for i, w := range text {
		if unicode.IsLower(w) && i == 0 {
			c++
			continue
		}
		if unicode.IsUpper(w) && (!isLastSymbolUpper || i == 0) {
			c++
			isLastSymbolUpper = true
			continue
		}
		isLastSymbolUpper = false
	}
	return c
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
	case *ast.Ident:
		return n.Name
	}
	return ""
}

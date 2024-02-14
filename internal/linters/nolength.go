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
	messageNoLength  = "The maximum length of object must be"
	messageNoSegment = "The maximum count segment of object must be"
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

type DrBigAlseFunctionVeryVeryTratatata struct {
	DrBigAlseFunctionVeryVeryField string
}

func runNoLength(pkgFiles []*ast.File, _ *types.Info, fset *token.FileSet) []Issue {
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.GenDecl)(nil),
	}

	inspect := inspector.New(pkgFiles)

	var pkgIssues []Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		switch n := node.(type) {
		case *ast.GenDecl:
			for _, spec := range n.Specs {
				nType, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}

				if nType.Name == nil {
					continue
				}

				if len(nType.Name.Name) > MaxLengthObject {
					position := fset.Position(node.Pos())

					pkgIssues = append(pkgIssues, Issue{
						Message:  fmt.Sprintf("%s %d (now `%s` %d)", messageNoLength, MaxLengthObject, nType.Name.Name, len(nType.Name.Name)),
						Line:     position.Line,
						Filename: position.Filename,
						Hash:     analysis.GetHashFromString(nType.Name.Name),
					})
				}

				segment := GetSegmentCount(nType.Name.Name)
				if segment > MaxSegmentCount {
					position := fset.Position(node.Pos())

					pkgIssues = append(pkgIssues, Issue{
						Message:  fmt.Sprintf("%s %d (now `%s` %d)", messageNoSegment, MaxSegmentCount, nType.Name.Name, segment),
						Line:     position.Line,
						Filename: position.Filename,
						Hash:     analysis.GetHashFromString(nType.Name.Name),
					})
				}

				nFunc, ok := nType.Type.(*ast.StructType)
				if !ok {
					continue
				}

				if nFunc.Fields == nil {
					continue
				}

				if nFunc.Fields.NumFields() == 0 {
					continue
				}

				for _, field := range nFunc.Fields.List {
					if len(field.Names) == 0 {
						continue
					}

					fieldName := field.Names[0].Name
					if len(fieldName) > MaxLengthObject {
						position := fset.Position(field.Pos())
						pkgIssues = append(pkgIssues, Issue{
							Message:  fmt.Sprintf("%s %d (now `%s` %d)", messageNoLength, MaxLengthObject, fieldName, len(fieldName)),
							Line:     position.Line,
							Filename: position.Filename,
							Hash:     analysis.GetHashFromString(fieldName),
						})
					}

					segment := GetSegmentCount(fieldName)
					if segment > MaxSegmentCount {
						position := fset.Position(field.Pos())

						pkgIssues = append(pkgIssues, Issue{
							Message:  fmt.Sprintf("%s %d (now `%s` %d)", messageNoSegment, MaxSegmentCount, fieldName, segment),
							Line:     position.Line,
							Filename: position.Filename,
							Hash:     analysis.GetHashFromString(fieldName),
						})
					}
				}
			}
		case *ast.FuncDecl:
			if n.Name == nil {
				return
			}

			if len(n.Name.Name) > MaxLengthObject {
				position := fset.Position(node.Pos())

				pkgIssues = append(pkgIssues, Issue{
					Message:  fmt.Sprintf("%s %d (now `%s` %d)", messageNoLength, MaxLengthObject, n.Name.Name, len(n.Name.Name)),
					Line:     position.Line,
					Filename: position.Filename,
					Hash:     analysis.GetHashFromString(n.Name.Name),
				})
			}

			segment := GetSegmentCount(n.Name.Name)
			if segment > MaxSegmentCount {
				position := fset.Position(node.Pos())

				pkgIssues = append(pkgIssues, Issue{
					Message:  fmt.Sprintf("%s %d (now `%s` %d)", messageNoSegment, MaxSegmentCount, n.Name.Name, segment),
					Line:     position.Line,
					Filename: position.Filename,
					Hash:     analysis.GetHashFromString(n.Name.Name),
				})
			}
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

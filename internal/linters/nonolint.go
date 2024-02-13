package linters

import (
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/mirecl/golimiter/internal/analysis"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoNoLint = "a `nolint` comment forbidden to use"
)

// NewNoNoLint create instance linter for check func nolint.
func NewNoNoLint() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoNoLint",
		Run: func(pkgs []*packages.Package) []Issue {
			issues := make([]Issue, 0)

			for _, p := range pkgs {
				pkgIssues := runNoNoLint(p.Syntax, p.TypesInfo, p.Fset)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoNoLint(pkgFiles []*ast.File, _ *types.Info, fset *token.FileSet) []Issue { //nolint:
	comments := make(map[string][]*ast.CommentGroup, len(pkgFiles))
	for _, file := range pkgFiles {
		comments[fset.Position(file.Pos()).Filename] = file.Comments
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	ignoreObjects := GetIgnore(pkgFiles, fset)

	inspect := inspector.New(pkgFiles)

	var pkgIssues []Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		nFuncDecl, _ := node.(*ast.FuncDecl)

		file := fset.Position(node.Pos()).Filename
		commentsFunc := GetCommentsByFunc(nFuncDecl, comments[file], fset)
		for _, comment := range commentsFunc {
			if !strings.Contains(comment.Text, "nolint:") {
				continue
			}

			if ignoreObjects.IsCheck(node) {
				continue
			}

			pkgIssues = append(pkgIssues, Issue{
				Message:  messageNoNoLint,
				Line:     comment.Line,
				Filename: comment.Filename,
				Hash:     analysis.GetHash(fset, node.Pos(), node.End()),
			})

		}
	})

	return pkgIssues
}

type FuncComment struct {
	Text     string
	Line     int
	Filename string
}

func GetCommentsByFunc(fn *ast.FuncDecl, fileComments []*ast.CommentGroup, fset *token.FileSet) []FuncComment {
	var comments []FuncComment

	for _, comment := range fileComments {
		if fn.Body.Pos() <= comment.Pos() && comment.Pos() <= fn.Body.End() {
			position := fset.Position(comment.Pos())
			comments = append(comments, FuncComment{
				Text:     comment.Text(),
				Line:     position.Line,
				Filename: position.Filename,
			})
		}
	}

	if fn.Doc != nil {
		for _, comment := range fn.Doc.List {
			position := fset.Position(comment.Pos())
			comments = append(comments, FuncComment{
				Text:     comment.Text,
				Line:     position.Line,
				Filename: position.Filename,
			})
		}
	}

	return comments
}

package linters

import (
	"go/ast"
	"go/token"
	"go/types"
	"regexp"
	"strings"

	"github.com/mirecl/golimiter/internal/analysis"
	"github.com/mirecl/golimiter/internal/config"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoNoLint = "a `nolint` comment forbidden to use"
)

var noLintRe = regexp.MustCompile(`nolint:(.*?)(\s|$)`)

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

// TODO: check nolint in struct
func runNoNoLint(pkgFiles []*ast.File, _ *types.Info, fset *token.FileSet) []Issue {
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
			b := noLintRe.FindStringSubmatch(comment.Text)
			if len(b) == 0 {
				continue
			}

			if nFuncDecl.Name != nil {
				lintres := strings.Split(b[1], ",")
				if config.Config.IsCheckNoLint(comment.Filename, nFuncDecl.Name.Name, lintres) {
					continue
				}
			}

			hash := analysis.GetHashFromPosition(fset, node)
			if ignoreObjects.IsCheck(hash) {
				continue
			}

			pkgIssues = append(pkgIssues, Issue{
				Message:  messageNoNoLint,
				Line:     comment.Line,
				Filename: comment.Filename,
				Hash:     hash,
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
		for _, c := range comment.List {
			if fn.Body.Pos() <= c.Pos() && c.Pos() <= fn.Body.End() {
				position := fset.Position(c.Pos())
				comments = append(comments, FuncComment{
					Text:     c.Text,
					Line:     position.Line,
					Filename: position.Filename,
				})
			}
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

package linters

import (
	"bufio"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/mirecl/golimiter/analysis"
	"github.com/mirecl/golimiter/config"
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
		Run: func(cfg *config.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoNoLint.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				issues = append(issues, runNoNoLint(&cfg.NoNoLint, pkg)...)
			}

			return issues
		},
	}
}

// TODO: check nolint in struct.
func runNoNoLint(cfg *config.NoNoLint, pkg *packages.Package) []analysis.Issue {
	comments := make(map[string][]*ast.CommentGroup, len(pkg.Syntax))
	for _, file := range pkg.Syntax {
		comments[pkg.Fset.Position(file.Pos()).Filename] = file.Comments
	}

	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
	}

	inspect := inspector.New(pkg.Syntax)

	var pkgIssues []analysis.Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		var hash string

		position := pkg.Fset.Position(node.Pos())

		currentFile := analysis.GetPathRelative(position.Filename)
		if slices.Contains(cfg.ExcludeFiles, currentFile) {
			return
		}

		nFuncDecl, _ := node.(*ast.FuncDecl)

		file := pkg.Fset.Position(node.Pos()).Filename
		commentsFunc := GetCommentsByFunc(nFuncDecl, comments[file], pkg.Fset)
		for _, comment := range commentsFunc {
			res := noLintRe.FindStringSubmatch(comment.Text)
			if len(res) == 0 {
				continue
			}

			if nFuncDecl.Name != nil {
				linters := strings.Split(res[1], ",")
				if cfg.IsVerifyName(comment.Filename, nFuncDecl.Name.Name, linters) {
					continue
				}
			}

			if comment.IsDoc {
				hash = analysis.GetHashFromBody(pkg.Fset, node)
			} else {
				hash = analysis.GetHashFromBodyByLine(pkg.Fset, node, comment.Line)
			}

			if cfg.IsVerifyHash(hash) {
				continue
			}

			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  messageNoNoLint,
				Line:     comment.Line,
				Filename: comment.Filename,
				Hash:     hash,
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
		}
	})

	return pkgIssues
}

type FuncComment struct {
	Text     string
	Line     int
	Filename string
	IsDoc    bool
}

func GetCommentsByFunc(fn *ast.FuncDecl, fileComments []*ast.CommentGroup, fset *token.FileSet) []FuncComment {
	var comments []FuncComment

	for _, comment := range fileComments {
		for _, c := range comment.List {
			if fn.Body == nil {
				continue
			}
			if fn.Body.Pos() <= c.Pos() && c.Pos() <= fn.Body.End() {
				position := fset.Position(c.Pos())
				comments = append(comments, FuncComment{
					Text:     c.Text,
					Line:     position.Line,
					Filename: position.Filename,
					IsDoc:    false,
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
				IsDoc:    true,
			})
		}
	}

	return comments
}

func ReadLine(path string, line int) string {
	file, err := os.Open(filepath.Clean(path))
	if err != nil {
		return ""
	}

	defer file.Close() //nolint:errcheck

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		i++
		if i != line {
			continue
		}
		return strings.TrimSpace(scanner.Text())
	}
	return ""
}

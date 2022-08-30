package exprcheck

import (
	"go/ast"
	"go/token"

	"github.com/mirecl/golimiter/internal"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Config linter.
type Config struct {
	Complexity int
}

// New instance linter.
func New(c *Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "exprcheck",
		Doc:      "Check complexity `expr` statement.",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return run(pass, c)
		},
	}
}

func run(pass *analysis.Pass, c *Config) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.IfStmt)(nil),
		(*ast.ReturnStmt)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		// check `*_test` files.
		if internal.IsTestFile(pass, node.Pos()) {
			return
		}

		r := result{}

		switch node := node.(type) {
		case *ast.IfStmt:
			check(node.Cond, r.collect)
		case *ast.ReturnStmt:
			for _, stmt := range node.Results {
				check(stmt, r.collect)
			}
		}

		if r.Complexity > c.Complexity {
			pass.Reportf(node.Pos(), "a complexity expr %d, allowed %d.", r.Complexity, c.Complexity)
		}
	})

	return nil, nil
}

type result struct {
	Complexity int
}

func (r *result) collect(t token.Token) {
	r.Complexity++
}

func check(expr ast.Expr, f func(t token.Token)) {
	switch n := expr.(type) {
	case *ast.BinaryExpr:
		f(n.Op)
		check(n.X, f)
		check(n.Y, f)
	case *ast.ParenExpr:
		check(n.X, f)
	}
}

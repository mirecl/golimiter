package exprcheck

import (
	"go/ast"

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
			r.check(node.Cond)
		case *ast.ReturnStmt:
			for _, stmt := range node.Results {
				r.check(stmt)
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

func (r *result) check(expr ast.Expr) {
	switch n := expr.(type) {
	case *ast.BinaryExpr:
		r.Complexity++
		r.check(n.X)
		r.check(n.Y)
	case *ast.ParenExpr:
		r.check(n.X)
	}
}

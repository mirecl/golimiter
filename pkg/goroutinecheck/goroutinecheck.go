package goroutinecheck

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Config linter.
type Config struct {
	Allowed *int
}

// Issues save global state.
var Issues []token.Pos

// New instance linter.
func New(c *Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "goroutinecheck",
		Doc:      "Check count `goroutine` statment.",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return run(c, pass)
		},
	}
}

func run(c *Config, pass *analysis.Pass) (interface{}, error) {
	// no restrictions.
	if c.Allowed == nil {
		return nil, nil
	}

	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.GoStmt)(nil)}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		Issues = append(Issues, node.Pos())
	})

	// check forbidden to use goroutine statment in all files.
	if *c.Allowed == 0 {
		for _, pos := range Issues {
			pass.Reportf(pos, "a init funcs forbidden to use.")
		}
		return nil, nil
	}

	// check forbidden to use  goroutine statment in all files if allowed num (Config) < issues.
	if *c.Allowed >= len(Issues) {
		return nil, nil
	}

	for _, pos := range Issues {
		pass.Reportf(pos, "a goroutine statment forbidden to use, but allowed %d", *c.Allowed)
	}

	return nil, nil
}

package goroutinecheck

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

type Config struct {
	Allowed int
}

var Issues []token.Pos

func New(cfg *Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "goroutinecheck",
		Doc:      "Check count `goroutine` statment.",
		Run:      cfg.run,
		Requires: []*analysis.Analyzer{inspect.Analyzer},
	}
}

func (c Config) run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.GoStmt)(nil)}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		Issues = append(Issues, node.Pos())
	})

	if c.Allowed >= len(Issues) {
		return nil, nil
	}

	for _, pos := range Issues {
		pass.Reportf(pos, "a goroutine statment forbidden to use.")
	}

	return nil, nil
}

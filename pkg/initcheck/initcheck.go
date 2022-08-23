package initcheck

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

const funcName = "init"

// Config linter.
type Config struct {
	Allowed *int
}

// Issues save global state.
var Issues []token.Pos

// New instance linter.
func New(c *Config) *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "initcheck",
		Doc:      "Check count `init` func.",
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

	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	var issues []token.Pos

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		fn, _ := node.(*ast.FuncDecl)

		if fn.Name.String() == funcName {
			issues = append(issues, node.Pos())
		}
	})

	Issues = append(Issues, issues...)

	// check forbidden to use `init` func in all files.
	if *c.Allowed == 0 {
		for _, pos := range issues {
			pass.Reportf(pos, "a init funcs forbidden to use.")
		}
		return nil, nil
	}

	// check forbidden to use `init` func in all files if allowed num (Config) < issues.
	if len(Issues) > *c.Allowed {
		for _, pos := range Issues {
			pass.Reportf(pos, "a found %d init funcs, but allowed %d.", len(Issues), *c.Allowed)
		}
		return nil, nil
	}

	// check forbidden to use `init` func in package if found > 1.
	if len(issues) > 1 {
		for _, pos := range issues {
			pass.Reportf(pos, "a found %d init funcs in package `%s`, but allowed 1.", len(issues), pass.Pkg.Name())
		}
		return nil, nil
	}

	return nil, nil
}

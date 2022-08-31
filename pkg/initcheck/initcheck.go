package initcheck

import (
	"go/ast"

	"github.com/mirecl/golimiter/internal"
	"github.com/mirecl/golimiter/internal/store"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Config linter.
type Config struct {
	Limit *int `yaml:"limit"`
}

// global state issues.
var state store.Store

// New instance linter.
func New(c *Config) *analysis.Analyzer {
	state = store.New()

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
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	var pkgIssues []*store.Issue

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		fileName := pass.Fset.Position(node.Pos()).Filename

		// check `*_test` files.
		if internal.IsTestFile(fileName) {
			return
		}

		fn, _ := node.(*ast.FuncDecl)

		if fn.Name.String() == "init" {
			pkgIssues = append(pkgIssues, &store.Issue{
				Pos:  node.Pos(),
				Pass: pass,
			})
		}
	})

	// forbidden all `init` funcs.
	if c.Limit != nil && *c.Limit == 0 {
		for _, issue := range pkgIssues {
			issue.Report("a `init` funcs forbidden to use.")
		}
		return nil, nil
	}

	// 1 `init` func in package - Ok.
	if len(pkgIssues) == 1 {
		state.Add(pkgIssues[0])
	}

	// 2 or more `init` func in package.
	if len(pkgIssues) > 1 {
		for _, issue := range pkgIssues {
			issue.Reportf("a found %d `init` funcs in package `%s`, but allowed only 1.", len(pkgIssues), pass.Pkg.Name())
			state.Add(issue)
		}
	}

	// limit `init` funcs.
	if c.Limit != nil && state.Len() > *c.Limit {
		state.Reportf("a number of allowed `init` funcs %d.", *c.Limit)
		return nil, nil
	}

	return nil, nil
}

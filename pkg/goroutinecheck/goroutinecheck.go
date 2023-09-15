package goroutinecheck

import (
	"go/ast"

	"github.com/mirecl/golimiter/internal"
	"github.com/mirecl/golimiter/internal/store"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// Exclude alias for internal type Exclude.
type Exclude = internal.Exclude

// Config linter.
type Config struct {
	Limit   *int `json:"limit" yaml:"limit"`
	Exclude `json:"exclude" yaml:"exclude"`
}

// global state issues.
var state store.Store

// New instance linter.
func New(c *Config) *analysis.Analyzer {
	state = store.New()

	return &analysis.Analyzer{
		Name:     "GoroutineCheck",
		Doc:      "Check count `goroutine` statement.",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return run(c, pass)
		},
	}
}

func run(c *Config, pass *analysis.Pass) (interface{}, error) {
	// no restrictions.
	if c.Limit == nil {
		return nil, nil
	}

	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{(*ast.GoStmt)(nil)}

	var pkgIssues []*store.Issue

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		if c.IsExclude(pass, node) {
			return
		}

		pkgIssues = append(pkgIssues, &store.Issue{
			Pos:  node.Pos(),
			Pass: pass,
		})
	})

	// forbidden all `goroutine` statement.
	if *c.Limit == 0 {
		for _, issue := range pkgIssues {
			issue.Report("a `goroutine` statement forbidden to use.")
		}
		return nil, nil
	}

	for _, issue := range pkgIssues {
		state.Add(issue)
	}

	// limit `goroutine` statement.
	if state.Len() >= *c.Limit {
		state.Reportf("a number of allowed `goroutine` statement %d.", *c.Limit)
		return nil, nil
	}

	return nil, nil
}

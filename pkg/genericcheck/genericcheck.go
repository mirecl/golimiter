package genericcheck

import (
	"go/ast"
	"go/types"

	"github.com/mirecl/golimiter/internal"
	"github.com/mirecl/golimiter/internal/store"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// New instance linter.
// Please more info in https://cs.opensource.google/go/x/tools/+/master:go/analysis/passes/usesgenerics/
func New() *analysis.Analyzer {
	return &analysis.Analyzer{
		Name:     "genericcheck",
		Doc:      "Check `generic` statement.",
		Requires: []*analysis.Analyzer{inspect.Analyzer},
		Run: func(pass *analysis.Pass) (interface{}, error) {
			return run(pass)
		},
	}
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.FuncType)(nil),
		(*ast.InterfaceType)(nil),
		(*ast.TypeSpec)(nil),
	}

	var pkgIssues []*store.Issue

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		var isGeneric bool

		// check `*_test` files.
		if internal.IsTestFile(pass, node.Pos()) {
			return
		}

		switch n := node.(type) {
		case *ast.FuncType:
			if tparams := ForFuncType(n); tparams != nil {
				isGeneric = true
			}
		case *ast.InterfaceType:
			tv := pass.TypesInfo.Types[n]
			if iface, _ := tv.Type.(*types.Interface); iface != nil && !iface.IsMethodSet() {
				isGeneric = true
			}
		case *ast.TypeSpec:
			if tparams := ForTypeSpec(n); tparams != nil {
				isGeneric = true
			}
		}

		if isGeneric {
			pkgIssues = append(pkgIssues, &store.Issue{
				Pos:  node.Pos(),
				Pass: pass,
			})
		}
	})

	for _, issue := range pkgIssues {
		issue.Report("a `generic` statement forbidden to use.")
	}

	return nil, nil
}

// ForTypeSpec returns TypeParams.
func ForTypeSpec(n *ast.TypeSpec) *ast.FieldList {
	if n == nil {
		return nil
	}
	return n.TypeParams
}

// ForFuncType returns TypeParams.
func ForFuncType(n *ast.FuncType) *ast.FieldList {
	if n == nil {
		return nil
	}
	return n.TypeParams
}

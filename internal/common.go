package internal

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// ExcludeType consist object for exclude from analysis.
type ExcludeType []string

// ConsistOf check object on excluded.
func (e ExcludeType) ConsistOf(fileName string) bool {
	if fileName == "" {
		return false
	}

	for _, value := range e {
		if strings.HasSuffix(fileName, value) {
			return true
		}
	}

	return false
}

// IsTestFile check test file or not.
func IsTestFile(fileName string) bool {
	return strings.HasSuffix(fileName, "_test.go")
}

// GetFuncDecl returns *ast.FuncDecl at position in source ccode.
func GetFuncDecl(pos token.Position) *ast.FuncDecl {
	file, err := os.Open(pos.Filename)
	if err != nil {
		return nil
	}
	defer file.Close() //nolint:errcheck,gosec

	fset := token.NewFileSet()

	astFile, err := parser.ParseFile(fset, "", file, 0)
	if err != nil {
		return nil
	}

	var fn *ast.FuncDecl

	ast.Inspect(astFile, func(node ast.Node) bool {
		funcDecl, ok := node.(*ast.FuncDecl)
		if !ok {
			return true
		}

		start := fset.Position(node.Pos())
		end := fset.Position(node.End())

		if pos.Line < start.Line {
			return true
		}

		if pos.Line > end.Line {
			return true
		}

		fn = funcDecl
		return true
	})

	return fn
}

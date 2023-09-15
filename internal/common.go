package internal

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis"
)

// ExcludeType consist object for exclude from analysis.
type ExcludeType []string

// ConsistOf check object on excluded.
func (e ExcludeType) ConsistOf(object string) bool {
	if object == "" {
		return false
	}

	for _, value := range e {
		if strings.HasSuffix(object, value) {
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

// Exclude rules.
type Exclude struct {
	ModFile *modfile.File
	Files   ExcludeType `json:"files" yaml:"files"`
	Funcs   ExcludeType `json:"funcs" yaml:"funcs"`
}

// IsExclude check all contidion for exclude analysis.
func (e Exclude) IsExclude(pass *analysis.Pass, node ast.Node) bool {
	folder, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	var fileName, fileMod, funcName, funcMod string

	pos := pass.Fset.Position(node.Pos())

	fileName = pos.Filename
	funcName = pass.Pkg.Name() + "." + GetFuncDecl(pos).Name.Name

	if e.ModFile != nil {
		fileMod = strings.ReplaceAll(fileName, folder, e.ModFile.Module.Mod.Path)
		funcMod = strings.ReplaceAll(filepath.Dir(fileName), folder, e.ModFile.Module.Mod.Path) + "/" + funcName
	}

	// check `*_test` files.
	if IsTestFile(fileName) {
		return true
	}

	// check exclude files.
	if e.Files.ConsistOf(fileName) || e.Files.ConsistOf(fileMod) {
		return true
	}

	// check exclude funcs.
	if e.Funcs.ConsistOf(funcName) || e.Funcs.ConsistOf(funcMod) {
		return true
	}

	return false
}

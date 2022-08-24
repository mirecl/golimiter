package internal

import (
	"go/token"
	"strings"

	"golang.org/x/tools/go/analysis"
)

func IsTestFile(pass *analysis.Pass, pos token.Pos) bool {
	fileName := pass.Fset.Position(pos).Filename
	return strings.HasSuffix(fileName, "_test.go")
}

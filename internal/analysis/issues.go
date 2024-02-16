package analysis

import (
	"crypto/md5"
	"encoding/hex"
	"go/ast"
	"go/token"
	"os"
)

// Issue problem in analysis.
type Issue struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Hash     string `json:"hash"`
}

// func (i Issue) Position() string {
// 	position := i.Fset.Position(i.Pos).String()

// 	dir, err := os.Getwd()
// 	if err != nil {
// 		return position
// 	}

// 	p, err := filepath.Rel(dir, position)
// 	if err != nil {
// 		return position
// 	}

// 	return p
// }

func GetHashFromPosition(fset *token.FileSet, node ast.Node) string {
	b, err := os.ReadFile(fset.Position(node.Pos()).Filename)
	if err != nil {
		return "Unknown"
	}

	hash := md5.Sum(b[fset.Position(node.Pos()).Offset:fset.Position(node.End()).Offset])
	return hex.EncodeToString(hash[:])
}

func GetHashFromString(object string) string {
	hash := md5.Sum([]byte(object))
	return hex.EncodeToString(hash[:])
}

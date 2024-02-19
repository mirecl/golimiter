package analysis

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"go/ast"
	"go/token"
	"os"
	"strings"
)

// Issue problem in analysis.
type Issue struct {
	Message  string `json:"message"`
	Filename string `json:"filename"`
	Line     int    `json:"line"`
	Hash     string `json:"hash"`
}

func GetHashFromBody(fset *token.FileSet, node ast.Node) string {
	b, err := os.ReadFile(fset.Position(node.Pos()).Filename)
	if err != nil {
		return ""
	}

	body := b[fset.Position(node.Pos()).Offset:fset.Position(node.End()).Offset]
	return GetHashFromBytes(body)
}

func GetHashFromString(object string) string {
	return GetHashFromBytes([]byte(object))
}

func GetHashFromBytes(object []byte) string {
	hash := md5.Sum(object)
	return hex.EncodeToString(hash[:])
}

func GetHashFromBodyByLine(fset *token.FileSet, node ast.Node, line int) string {
	file, err := os.Open(fset.Position(node.Pos()).Filename)
	if err != nil {
		return ""
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	i := 0
	for scanner.Scan() {
		i++
		if i != line {
			continue
		}

		body := strings.TrimSpace(scanner.Text())
		return GetHashFromString(body)
	}

	return ""
}

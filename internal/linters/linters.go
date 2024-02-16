package linters

import (
	"crypto/md5"
	"encoding/hex"
	"go/ast"
	"go/token"
	"os"
	"regexp"

	"github.com/mirecl/golimiter/internal/analysis"
	"github.com/mirecl/golimiter/internal/config"
)

type Issue = analysis.Issue

var IgnoreRe = regexp.MustCompile(`/.*GOLIMITER:([A-Fa-f0-9]{32}).*`)

type Object struct {
	File string
	Line int
	Hash string
}

type IgnoreScope struct {
	objects []Object
	fset    *token.FileSet
}

func (i *IgnoreScope) IsCheck(hash string) bool {
	for _, object := range i.objects {
		if object.Hash == hash && config.Config.IsCheckHash(hash) {
			return true
		}
	}
	return false
}

func GetIgnore(pkgFiles []*ast.File, fset *token.FileSet) IgnoreScope {
	objects := make([]Object, 0)
	for _, c := range pkgFiles {
		for _, i := range c.Comments {
			for _, k := range i.List {
				b := IgnoreRe.FindStringSubmatch(k.Text)
				if len(b) != 2 {
					continue
				}

				position := fset.Position(i.End())
				objects = append(objects, Object{
					Line: position.Line,
					File: position.Filename,
					Hash: b[1],
				})
			}
		}
	}
	return IgnoreScope{
		objects: objects,
		fset:    fset,
	}
}

// Hash generate hash from body file.
func Hash(fset *token.FileSet, node ast.Node) string {
	b, err := os.ReadFile(fset.Position(node.Pos()).Filename)
	if err != nil {
		return "Unknown"
	}

	start := fset.Position(node.Pos()).Offset
	end := fset.Position(node.End()).Offset

	hash := md5.Sum(b[start:end])
	return hex.EncodeToString(hash[:])
}

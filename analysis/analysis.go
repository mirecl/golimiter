package analysis

import (
	"bufio"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"go/ast"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/mirecl/golimiter/config"
	"golang.org/x/tools/go/packages"
)

const loadMode = packages.NeedName |
	packages.NeedFiles |
	packages.NeedCompiledGoFiles |
	packages.NeedImports |
	packages.NeedTypes |
	packages.NeedTypesSizes |
	packages.NeedTypesInfo |
	packages.NeedSyntax |
	packages.NeedModule |
	packages.NeedEmbedFiles |
	packages.NeedEmbedPatterns

// Linter describes an analysis function and its options.
type Linter struct {
	// The Name of the analyzer must be a valid Go identifier
	Name string
	// Run applies the analyzer to a package.
	Run func(*config.Config, []*packages.Package) []Issue
}

// Run analyze source code.
func Run(cfg *config.Config, linters ...*Linter) map[string][]Issue {
	pkgs, err := packages.Load(&packages.Config{Mode: loadMode, Tests: false}, "./...")
	if err != nil {
		log.Fatalf("failed load go/packages: %s", err)
	}

	allIssues := make(map[string][]Issue, len(linters))

	for _, linter := range linters {
		allIssues[linter.Name] = linter.Run(cfg, pkgs)
	}

	return allIssues
}

func GetHashFromBody(fset *token.FileSet, node ast.Node) string {
	filename := fset.Position(node.Pos()).Filename
	filename = GetPathRelative(filename)

	b, err := os.ReadFile(filename)
	if err != nil {
		return ""
	}

	body := b[fset.Position(node.Pos()).Offset:fset.Position(node.End()).Offset]
	body = append(body, []byte(filename)...)
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
	filename := fset.Position(node.Pos()).Filename
	filename = GetPathRelative(filename)

	file, err := os.Open(filename)
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

		body := fmt.Sprintf("%s_%s", strings.TrimSpace(scanner.Text()), filename)
		return GetHashFromString(body)
	}

	return ""
}

func GetPathRelative(path string) string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	path, err = filepath.Rel(dir, path)
	if err != nil {
		panic(err)
	}
	return path
}

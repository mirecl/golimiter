package linters

import (
	"fmt"
	"go/ast"
	"strings"
	"unicode"

	"github.com/mirecl/golimiter/analysis"
	"github.com/mirecl/golimiter/config"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoDocTag = "in the struct `%s`, the field `%s` does not have a required tag `doc`"
)

// NewNoDoc create instance linter for check docs.
func NewNoDoc() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoDoc",
		Run: func(cfg *config.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoDoc.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				pkgIssues := runNoDocTag(&cfg.NoDoc, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoDocTag(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	var pkgIssues []analysis.Issue

	gomodfile, err := config.ReadModFile()
	if err != nil {
		panic(err)
	}

	pkgName := strings.ReplaceAll(pkg.PkgPath, fmt.Sprintf("%s/", gomodfile.Module.Mod.Path), "")

	if !strings.HasPrefix(pkgName, "pkg") {
		return pkgIssues
	}

	nodeFilter := []ast.Node{(*ast.TypeSpec)(nil)}
	inspect := inspector.New(pkg.Syntax)

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		typeSpec := node.(*ast.TypeSpec)

		// only proceed with struct types
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return
		}

		// get type name
		if typeSpec.Name == nil {
			return
		}
		typeName := typeSpec.Name.Name

		if !unicode.IsUpper(rune(typeName[0])) {
			return
		}

		for _, field := range structType.Fields.List {
			position := pkg.Fset.Position(field.Pos())

			if len(field.Names) == 0 {
				continue
			}
			filedName := field.Names[0].String()

			if field.Tag != nil && strings.Contains(field.Tag.Value, "doc:") {
				continue
			}

			hash := analysis.GetHashFromString(fmt.Sprintf("%s.%s", typeName, filedName))
			if cfg.IsVerifyHash(hash) {
				continue
			}

			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  fmt.Sprintf(messageNoDocTag, typeName, filedName),
				Line:     position.Line,
				Filename: position.Filename,
				Hash:     hash,
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
		}
	})

	return pkgIssues
}

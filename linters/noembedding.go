package linters

import (
	"fmt"
	"go/ast"
	"go/token"
	"slices"
	"strings"

	"github.com/mirecl/golimiter/analysis"
	"github.com/mirecl/golimiter/config"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoStructEmbedding = "a `embedding` struct forbidden to use - field name `%s`"
)

// NewEmbedding create instance linter for check embedding.
func NewEmbedding() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoEmbedding",
		Run: func(cfg *config.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoEmbedding.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				pkgIssues := runNoStructEmbedding(&cfg.NoEmbedding, pkg)
				issues = append(issues, pkgIssues...)
			}

			return issues
		},
	}
}

func runNoStructEmbedding(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	nodeFilter := []ast.Node{(*ast.TypeSpec)(nil)}

	inspect := inspector.New(pkg.Syntax)

	var pkgIssues []analysis.Issue

	gomodfile, err := config.ReadModFile()
	if err != nil {
		panic(err)
	}

	pkgName := strings.ReplaceAll(pkg.PkgPath, fmt.Sprintf("%s/", gomodfile.Module.Mod.Path), "")

	if !strings.HasPrefix(pkgName, "pkg/request") && !strings.HasPrefix(pkgName, "pkg/response") {
		return pkgIssues
	}

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		typeSpec := node.(*ast.TypeSpec)

		// only proceed with struct types
		structType, ok := typeSpec.Type.(*ast.StructType)
		if !ok {
			return
		}

		position := pkg.Fset.Position(node.Pos())

		currentFile := analysis.GetPathRelative(position.Filename)
		if slices.Contains(cfg.ExcludeFiles, currentFile) {
			return
		}

		for _, folder := range cfg.ExcludeFolders {
			if strings.HasPrefix(currentFile, folder) {
				return
			}
		}

		for _, field := range findEmbeddedFields(structType) {
			p := pkg.Fset.Position(field.Pos)

			hash := analysis.GetHashFromString(p.Filename + field.Name + typeSpec.Name.String())
			if cfg.IsVerifyHash(hash) {
				continue
			}

			if strings.Contains(field.Tag, "json:") {
				continue
			}

			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  fmt.Sprintf(messageNoStructEmbedding, field.Name),
				Line:     p.Line,
				Filename: p.Filename,
				Hash:     hash,
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
		}
	})

	return pkgIssues
}

// EmbeddedFields info about embedded fields.
type EmbeddedFields struct {
	Name string
	Pos  token.Pos
	Tag  string
}

func findEmbeddedFields(st *ast.StructType) []EmbeddedFields {
	var embeddedFields []EmbeddedFields
	for _, field := range st.Fields.List {
		// анонимное поле
		if len(field.Names) != 0 {
			continue
		}

		tag := ""
		if field.Tag != nil {
			tag = field.Tag.Value
		}

		// Получаем тип поля
		fieldType := field.Type
		switch t := fieldType.(type) {
		case *ast.Ident: // Простое имя
			embeddedFields = append(embeddedFields, EmbeddedFields{t.Name, t.Pos(), tag})
		case *ast.SelectorExpr: // Тип с пакетом, например pkg.T
			embeddedFields = append(embeddedFields, EmbeddedFields{t.Sel.Name, t.Pos(), tag})
		}
	}

	return embeddedFields
}

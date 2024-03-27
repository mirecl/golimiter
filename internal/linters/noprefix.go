package linters

import (
	"fmt"
	"go/ast"
	"slices"
	"strings"

	"github.com/mirecl/golimiter/config"
	"github.com/mirecl/golimiter/internal/analysis"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoPrefix       = "a `prefix` funcs forbidden to use"
	messageNoPrefixLambda = "a `lambda` funcs forbidden to use"
)

var PrefixAllow = []string{"get", "new", "is", "calc", "validate", "normalize",
	"execute", "get", "set", "parse", "apply", "append", "clear", "remove",
	"delete", "update", "to", "from", "run", "read", "collect", "add", "predict",
	"inference", "check", "max", "min", "find"}

// NewNoInit create instance linter for check func init.
//
//nolint:dupl
func NewNoPrefix() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoPrefix",
		Run: func(cfg *config.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoPrefix.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				issues = append(issues, runNoPrefix(&cfg.NoPrefix, pkg)...)
				issues = append(issues, runNoCommonPrefix(&cfg.NoPrefix, pkg)...)
			}

			return issues
		},
	}
}

func runNoPrefix(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}

	inspect := inspector.New(pkg.Syntax)

	var pkgIssues []analysis.Issue

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		fn, _ := node.(*ast.FuncDecl)
		position := pkg.Fset.Position(node.Pos())

		currentFile := analysis.GetPathRelative(position.Filename)
		if slices.Contains(cfg.ExcludeFiles, currentFile) {
			return
		}

		if fn.Name == nil {
			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  messageNoPrefixLambda,
				Line:     position.Line,
				Filename: position.Filename,
				Hash:     "",
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
			return
		}

		name := strings.ToLower(fn.Name.Name)

		if name == "main" || name == "init" {
			return
		}

		for _, prefix := range PrefixAllow {
			if strings.HasPrefix(name, prefix) {
				return
			}
		}

		hash := analysis.GetHashFromString(fn.Name.Name)
		if cfg.IsVerifyHash(hash) {
			return
		}

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  messageNoPrefix,
			Line:     position.Line,
			Filename: position.Filename,
			Hash:     hash,
			Severity: cfg.Severity,
			Type:     cfg.Type,
		})
	})

	return pkgIssues
}

func runNoCommonPrefix(cfg *config.DefaultLinter, pkg *packages.Package) (pkgIssues []analysis.Issue) {

	inspect := inspector.New(pkg.Syntax)

	nodeFilter := []ast.Node{(*ast.TypeSpec)(nil)}

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		position := pkg.Fset.Position(node.Pos())

		currentFile := analysis.GetPathRelative(position.Filename)
		if slices.Contains(cfg.ExcludeFiles, currentFile) {
			return
		}

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

		var (
			fieldNames []string               // field names
			fieldIdx   = make(map[string]int) // index of field in StructType.Fields.List
		)

		// collect all field names
		for i, field := range structType.Fields.List {
			if field.Names == nil {
				// embedded type
				continue
			}

			fieldName := field.Names[0].Name
			fieldNames = append(fieldNames, fieldName)
			fieldIdx[fieldName] = i
		}

		// find fields with common prefix
		commonPrefix, found := FindIdentsWithPartialPrefix(typeName, fieldNames)

		// make issues
		for _, fieldName := range found {
			hash := analysis.GetHashFromString(typeName + fieldName)
			if cfg.IsVerifyHash(hash) {
				continue
			}

			fieldPos := pkg.Fset.Position(structType.Fields.List[fieldIdx[fieldName]].Pos())

			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  fmt.Sprintf("field %s has common prefix (%s) with struct name (%s)", fieldName, commonPrefix, typeName),
				Line:     fieldPos.Line,
				Filename: fieldPos.Filename,
				Hash:     hash,
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
		}
	})

	return pkgIssues
}

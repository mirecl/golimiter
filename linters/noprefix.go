package linters

import (
	"fmt"
	"go/ast"
	"go/types"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/mirecl/golimiter/analysis"
	"github.com/mirecl/golimiter/config"
	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoPrefixLambda = "a `lambda` funcs forbidden to use"
)

var action = []string{"get", "new", "calc", "validate", "normalize",
	"execute", "set", "parse", "apply", "append", "clear", "remove",
	"delete", "update", "run", "read", "write", "collect", "add", "predict",
	"inference", "check", "max", "min", "find", "is", "any", "all"}

// NewNoPrefix create instance linter for check func prefix.
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

		name := fn.Name.Name

		if name == "main" || name == "init" {
			return
		}

		fixName := FixNameFromFuncDecl(fn)
		if fixName == fn.Name.Name {
			return
		}

		hash := analysis.GetHashFromString(name)
		if cfg.IsVerifyHash(hash) {
			return
		}

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  fmt.Sprintf("please rename func `%s` → `%s`", name, fixName),
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

func FixNameFromFuncDecl(fn *ast.FuncDecl) string {
	if fn.Name == nil {
		return ""
	}

	text := fn.Name.Name
	segmentes := GetSegments(text)

	if fn.Type.Results != nil {
		if len(fn.Type.Results.List) == 1 {
			returnType := types.ExprString(fn.Type.Results.List[0].Type)
			isAction := slices.Contains(action, strings.ToLower(segmentes[0]))

			if returnType == "bool" && !isAction {
				if IsLower(text[0]) {
					segmentes[0] = FirstToUpper(segmentes[0])
					segmentes = append([]string{"is"}, segmentes...)
				} else {
					segmentes = append([]string{"Is"}, segmentes...)
				}
			}
		}
	}

	fixName := strings.Join(segmentes, "")
	if fixName != text {
		return fixName
	}

	return FixName(text)
}

func FixName(text string) string {

	segmentes := GetSegments(text)
	res := make([]string, 0, len(segmentes))

	for i, segment := range segmentes {
		isAction := slices.Contains(action, strings.ToLower(segment))

		if i == 0 && isAction {
			return text
		}

		if isAction {
			if IsLower(text[0]) {
				segment = FirstToLower(segment)
				res[0] = FirstToUpper(res[0])
			}
			res = append([]string{segment}, res...)
			continue
		}
		res = append(res, segment)
	}

	return strings.Join(res, "")
}

func IsLower(b byte) bool {
	return unicode.IsLower(rune(b)) || !unicode.IsLetter(rune(b))
}

func FirstToLower(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}

func FirstToUpper(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToUpper(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}

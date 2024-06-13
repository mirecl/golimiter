package linters

import (
	"fmt"
	"go/ast"
	"go/token"
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
	messageNoPrefixLambda                   = "a `lambda` funcs forbidden to use"
	messageNoPrefixUpperFirstSymbolVariable = "please not use `%s` with first Upper symbol in variable"
	messageNoPrefixUpperFirstSymbolParams   = "please not use `%s` with first Upper symbol in params"
	messageNoPrefixUpperFirstSymbolReturns  = "please not use `%s` with first Upper symbol in returns"
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
				issues = append(issues, runNoPrefixUpperSymbol(&cfg.NoPrefix, pkg)...)
			}

			return issues
		},
	}
}

func GetParamsFromFunc(funcType *ast.FuncType) []string {
	params := funcType.Params
	if params.NumFields() == 0 {
		return nil
	}

	res := make([]string, 0, params.NumFields())

	for _, field := range params.List {
		if field.Names[0].Name != "_" {
			res = append(res, field.Names[0].Name)
		}
	}
	return res
}

func GetReturnsFromFunc(funcType *ast.FuncType) []string {
	returns := funcType.Results
	if returns.NumFields() == 0 {
		return nil
	}

	res := make([]string, 0, returns.NumFields())

	for _, field := range returns.List {
		if len(field.Names) == 0 {
			continue
		}
		if field.Names[0].Name != "_" && field.Names[0].Name != "" {
			res = append(res, field.Names[0].Name)
		}
	}
	return res
}

type BodyVariable struct {
	Name     string
	Position token.Pos
}

func GetVarFromBody(block *ast.BlockStmt) []BodyVariable {
	var variables []BodyVariable
	for _, stmt := range block.List {
		if assignStmt, ok := stmt.(*ast.AssignStmt); ok {
			if ident, ok := assignStmt.Lhs[0].(*ast.Ident); ok {
				fmt.Println(ident.Obj, ident.Obj.Data, ident.Name)
				if ident.Obj.Data != 0 {
					variables = append(variables, BodyVariable{
						Name:     ident.Name,
						Position: ident.NamePos,
					})
				}
			}
		}
		if declStmt, ok := stmt.(*ast.DeclStmt); ok {
			if genDecl, ok := declStmt.Decl.(*ast.GenDecl); ok {
				if genDecl.Tok == token.VAR {
					for _, spec := range genDecl.Specs {
						if valueSpec, ok := spec.(*ast.ValueSpec); ok {
							for _, ident := range valueSpec.Names {
								variables = append(variables, BodyVariable{
									Name:     ident.Name,
									Position: ident.NamePos,
								})
							}
						}
					}
				}
			}
		}
	}
	return variables
}

func runNoPrefixUpperSymbol(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	var pkgIssues []analysis.Issue

	nodeFilter := []ast.Node{(*ast.FuncDecl)(nil)}
	inspect := inspector.New(pkg.Syntax)

	inspect.Preorder(nodeFilter, func(node ast.Node) {
		decl := node.(*ast.FuncDecl)

		for _, field := range GetVarFromBody(decl.Body) {
			if !unicode.IsUpper(rune(field.Name[0])) {
				continue
			}

			position := pkg.Fset.Position(field.Position)

			currentFile := analysis.GetPathRelative(position.Filename)
			if slices.Contains(cfg.ExcludeFiles, currentFile) {
				return
			}

			for _, folder := range cfg.ExcludeFolders {
				if strings.HasPrefix(currentFile, folder) {
					return
				}
			}

			hash := analysis.GetHashFromString(field.Name)

			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  fmt.Sprintf(messageNoPrefixUpperFirstSymbolVariable, field.Name),
				Line:     position.Line,
				Filename: position.Filename,
				Hash:     hash,
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
		}

		position := pkg.Fset.Position(node.Pos())

		for _, field := range GetParamsFromFunc(decl.Type) {
			currentFile := analysis.GetPathRelative(position.Filename)
			if slices.Contains(cfg.ExcludeFiles, currentFile) {
				return
			}

			for _, folder := range cfg.ExcludeFolders {
				if strings.HasPrefix(currentFile, folder) {
					return
				}
			}

			if !unicode.IsUpper(rune(field[0])) {
				continue
			}

			hash := analysis.GetHashFromString(field)

			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  fmt.Sprintf(messageNoPrefixUpperFirstSymbolParams, field),
				Line:     position.Line,
				Filename: position.Filename,
				Hash:     hash,
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
		}

		for _, field := range GetReturnsFromFunc(decl.Type) {
			currentFile := analysis.GetPathRelative(position.Filename)
			if slices.Contains(cfg.ExcludeFiles, currentFile) {
				return
			}

			for _, folder := range cfg.ExcludeFolders {
				if strings.HasPrefix(currentFile, folder) {
					return
				}
			}

			if !unicode.IsUpper(rune(field[0])) {
				continue
			}

			hash := analysis.GetHashFromString(field)

			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  fmt.Sprintf(messageNoPrefixUpperFirstSymbolReturns, field),
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

		for _, folder := range cfg.ExcludeFolders {
			if strings.HasPrefix(currentFile, folder) {
				return
			}
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

		hash := analysis.GetHashFromString(name)
		if cfg.IsVerifyHash(hash) {
			return
		}

		fixName := FixNameFromFuncDecl(fn)
		if fixName == fn.Name.Name {
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

		for _, folder := range cfg.ExcludeFolders {
			if strings.HasPrefix(currentFile, folder) {
				return
			}
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

	text := strings.ReplaceAll(fn.Name.Name, "_", "")

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

	fixName := strings.TrimSpace(strings.Join(segmentes, ""))
	if fixName != text {
		return fixName
	}

	if len(segmentes) == 1 {
		return text
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

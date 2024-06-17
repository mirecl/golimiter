package linters

import (
	"fmt"
	"strings"

	"github.com/mirecl/golimiter/analysis"
	"github.com/mirecl/golimiter/config"
	"golang.org/x/tools/go/packages"
)

const (
	messageNoObjectPackageFile = "not found in package `%s` main file `%s.go`"
	messageNoObjectScripts     = "package with name `scripts` allowed to use only in root"
	messageNoObjectMain        = "file with name `main.go` allowed to use only in root"
)

func NewNoObject() *analysis.Linter {
	return &analysis.Linter{
		Name: "NoObject",
		Run: func(cfg *config.Config, pkgs []*packages.Package) []analysis.Issue {
			issues := make([]analysis.Issue, 0)

			if cfg.NoObject.Disable {
				return issues
			}

			for _, pkg := range pkgs {
				issues = append(issues, runNoObjectPackageFile(&cfg.NoObject, pkg)...)
				issues = append(issues, runNoObjectScripts(&cfg.NoObject, pkg)...)
				issues = append(issues, runNoObjectMainFile(&cfg.NoObject, pkg)...)
			}

			return issues
		},
	}
}

func runNoObjectMainFile(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	var pkgIssues []analysis.Issue

	gomodfile, err := config.ReadModFile()
	if err != nil {
		panic(err)
	}

	for _, file := range pkg.GoFiles {
		filePath := analysis.GetPathRelative(file)
		fileName := strings.ReplaceAll(filePath, fmt.Sprintf("%s/", gomodfile.Module.Mod.Path), "")

		if fileName == "main.go" {
			continue
		}

		if strings.HasSuffix(fileName, "main.go") {
			hash := analysis.GetHashFromString(file)
			if cfg.IsVerifyHash(hash) {
				return pkgIssues
			}

			pkgIssues = append(pkgIssues, analysis.Issue{
				Message:  messageNoObjectMain,
				Line:     1,
				Filename: file,
				Hash:     hash,
				Severity: cfg.Severity,
				Type:     cfg.Type,
			})
		}
	}

	return pkgIssues
}

func runNoObjectScripts(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	var pkgIssues []analysis.Issue

	gomodfile, err := config.ReadModFile()
	if err != nil {
		panic(err)
	}

	pkgName := strings.ReplaceAll(pkg.PkgPath, fmt.Sprintf("%s/", gomodfile.Module.Mod.Path), "")

	if pkgName == "scripts" {
		return pkgIssues
	}

	if !strings.Contains(pkgName, "scripts") {
		return pkgIssues
	}

	if len(pkg.GoFiles) == 0 {
		return pkgIssues
	}

	hash := analysis.GetHashFromString(pkg.PkgPath)
	if cfg.IsVerifyHash(hash) {
		return pkgIssues
	}

	pkgIssues = append(pkgIssues, analysis.Issue{
		Message:  messageNoObjectScripts,
		Line:     1,
		Filename: pkg.GoFiles[0],
		Hash:     hash,
		Severity: cfg.Severity,
		Type:     cfg.Type,
	})

	return pkgIssues
}

func runNoObjectPackageFile(cfg *config.DefaultLinter, pkg *packages.Package) []analysis.Issue {
	var pkgIssues []analysis.Issue

	isFind := false

	gomodfile, err := config.ReadModFile()
	if err != nil {
		panic(err)
	}

	pkgName := strings.ReplaceAll(pkg.PkgPath, fmt.Sprintf("%s/", gomodfile.Module.Mod.Path), "")
	if pkg.Name == "pkg" || pkgName == "scripts" {
		return pkgIssues
	}

L:
	for _, file := range pkg.GoFiles {
		if strings.Contains(file, fmt.Sprintf("%s.go", pkg.Name)) {
			isFind = true
			break L
		}
	}

	if !isFind {
		hash := analysis.GetHashFromString(fmt.Sprintf("%s_%s", pkg.PkgPath, pkg.Name))
		if cfg.IsVerifyHash(hash) {
			return pkgIssues
		}

		pkgIssues = append(pkgIssues, analysis.Issue{
			Message:  fmt.Sprintf(messageNoObjectPackageFile, pkg.PkgPath, pkg.Name),
			Line:     1,
			Filename: pkg.GoFiles[0],
			Hash:     hash,
			Severity: cfg.Severity,
			Type:     cfg.Type,
		})
	}

	return pkgIssues
}

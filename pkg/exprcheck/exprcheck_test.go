package exprcheck_test

import (
	"testing"

	"github.com/mirecl/golimiter/pkg/exprcheck"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis/analysistest"
)

var modBytes = []byte(`module github.com/repo/name

go 1.18
`)

func TestIf(t *testing.T) {
	testdata := analysistest.TestData()

	TestCases := []struct {
		cfg  *exprcheck.Config
		name string
		pkg  []string
	}{
		{
			name: "if-return, 1 func, expr allowed 0",
			pkg:  []string{"a"},
			cfg:  new(exprcheck.Config),
		},
		{
			name: "if, 2 funcs, expr allowed 0",
			pkg:  []string{"b"},
			cfg:  new(exprcheck.Config),
		},
		{
			name: "return, 2 funcs, expr allowed 0",
			pkg:  []string{"c"},
			cfg:  new(exprcheck.Config),
		},
		{
			name: "if-return, 1 funcs, expr allowed 3",
			pkg:  []string{"e"},
			cfg: func() *exprcheck.Config {
				complexity := 3
				return &exprcheck.Config{Complexity: complexity}
			}(),
		},
		{
			name: "exclude file, expr allowed 3",
			pkg:  []string{"f"},
			cfg: func() *exprcheck.Config {
				complexity := 3
				return &exprcheck.Config{
					Complexity:   complexity,
					ExcludeFiles: []string{"f.go"}}
			}(),
		},
		{
			name: "exclude func, expr allowed 3",
			pkg:  []string{"g"},
			cfg: func() *exprcheck.Config {
				complexity := 3
				return &exprcheck.Config{
					Complexity:   complexity,
					ExcludeFuncs: []string{"g.g2"}}
			}(),
		},
		{
			name: "exclude func/file, expr allowed 3",
			pkg:  []string{"h"},
			cfg: func() *exprcheck.Config {
				complexity := 3
				gomodfile, _ := modfile.Parse("go.mod", modBytes, nil)

				return &exprcheck.Config{
					Complexity: complexity,
					ModFile:    gomodfile,
					ExcludeFuncs: []string{
						"github.com/repo/name/testdata/src/h/h.h31",
						"github.com/repo/name/testdata/src/h/h.h2",
					},
					ExcludeFiles: []string{"github.com/repo/name/testdata/src/h/h1.go"}}
			}(),
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.name, func(t *testing.T) {
			analyzer := exprcheck.New(testCase.cfg)
			analysistest.Run(t, testdata, analyzer, testCase.pkg...)
		})
	}
}

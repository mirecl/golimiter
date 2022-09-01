package exprcheck_test

import (
	"testing"

	"github.com/mirecl/golimiter/internal"
	"github.com/mirecl/golimiter/pkg/exprcheck"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis/analysistest"
)

type Exclude = internal.Exclude

var modBytes = []byte(`module github.com/repo/name`)

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
				e := Exclude{Files: []string{"f.go"}}

				return &exprcheck.Config{
					Complexity: complexity,
					Exclude:    e,
				}
			}(),
		},
		{
			name: "exclude func, expr allowed 3",
			pkg:  []string{"g"},
			cfg: func() *exprcheck.Config {
				complexity := 3
				e := Exclude{Funcs: []string{"g.g2"}}

				return &exprcheck.Config{
					Complexity: complexity,
					Exclude:    e,
				}
			}(),
		},
		{
			name: "exclude func/file, expr allowed 3",
			pkg:  []string{"h"},
			cfg: func() *exprcheck.Config {
				complexity := 3
				ModFile, _ := modfile.Parse("go.mod", modBytes, nil)
				e := Exclude{
					ModFile: ModFile,
					Funcs: []string{
						"github.com/repo/name/testdata/src/h/h.h31",
						"github.com/repo/name/testdata/src/h/h.h2",
					},
					Files: []string{"github.com/repo/name/testdata/src/h/h1.go"},
				}

				return &exprcheck.Config{
					Complexity: complexity,
					Exclude:    e,
				}
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

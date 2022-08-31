package goroutinecheck_test

import (
	"testing"

	"github.com/mirecl/golimiter/pkg/goroutinecheck"
	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis/analysistest"
)

var modBytes = []byte(`module github.com/repo/name

go 1.18
`)

func TestGoroutine(t *testing.T) {
	testdata := analysistest.TestData()

	TestCases := []struct {
		cfg  *goroutinecheck.Config
		name string
		pkg  []string
	}{
		{
			name: "success analysis",
			pkg:  []string{"a"},
			cfg:  new(goroutinecheck.Config),
		},
		{
			name: "failed to 1 package",
			pkg:  []string{"b"},
			cfg: func() *goroutinecheck.Config {
				Limit := 2
				return &goroutinecheck.Config{Limit: &Limit}
			}(),
		},
		{
			name: "failed limit",
			pkg:  []string{"c"},
			cfg: func() *goroutinecheck.Config {
				Limit := 0
				return &goroutinecheck.Config{Limit: &Limit}
			}(),
		},
		{
			name: "failed limit in all package",
			pkg:  []string{"c", "d"},
			cfg: func() *goroutinecheck.Config {
				Limit := 0
				return &goroutinecheck.Config{Limit: &Limit}
			}(),
		},
		{
			name: "success analysis with test file",
			pkg:  []string{"e"},
			cfg: func() *goroutinecheck.Config {
				Limit := 10
				return &goroutinecheck.Config{Limit: &Limit}
			}(),
		},
		{
			name: "success analysis with exclude main",
			pkg:  []string{"f"},
			cfg: func() *goroutinecheck.Config {
				Limit := 0
				gomodfile, _ := modfile.Parse("go.mod", modBytes, nil)
				return &goroutinecheck.Config{
					Limit:        &Limit,
					ModFile:      gomodfile,
					ExcludeFuncs: []string{"f.main", "github.com/repo/name/testdata/src/f/f.main5"},
					ExcludeFiles: []string{"f/f3.go", "github.com/repo/name/testdata/src/f/f4.go"},
				}
			}(),
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.name, func(t *testing.T) {
			analyzer := goroutinecheck.New(testCase.cfg)
			analysistest.Run(t, testdata, analyzer, testCase.pkg...)
		})
	}
}

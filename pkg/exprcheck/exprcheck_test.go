package exprcheck_test

import (
	"testing"

	"github.com/mirecl/golimiter/pkg/exprcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestIf(t *testing.T) {
	testdata := analysistest.TestData()

	TestCases := []struct {
		cfg  *exprcheck.Config
		name string
		pkg  []string
	}{
		{
			name: "if/return, 1 func, expr allowed 0",
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
			name: "if/return, 1 funcs, expr allowed 3",
			pkg:  []string{"e"},
			cfg: func() *exprcheck.Config {
				complexity := 3
				return &exprcheck.Config{complexity}
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

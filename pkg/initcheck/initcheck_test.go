package initcheck_test

import (
	"testing"

	"github.com/mirecl/golimiter/pkg/initcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestInit(t *testing.T) {
	testdata := analysistest.TestData()

	TestCases := []struct {
		name string
		pkg  []string
		cfg  *initcheck.Config
	}{
		{
			name: "success analysis",
			pkg:  []string{"a"},
			cfg:  new(initcheck.Config),
		},
		{
			name: "forbidden to use",
			pkg:  []string{"b"},
			cfg: func() *initcheck.Config {
				Limit := 0
				return &initcheck.Config{Limit: &Limit}
			}(),
		},
		{
			name: "failed to 1 package",
			pkg:  []string{"c"},
			cfg:  new(initcheck.Config),
		},
		{
			name: "failed limit in all package",
			pkg:  []string{"c", "d"},
			cfg: func() *initcheck.Config {
				Limit := 2
				return &initcheck.Config{Limit: &Limit}
			}(),
		},
		{
			name: "success analysis with test file",
			pkg:  []string{"e"},
			cfg:  new(initcheck.Config),
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.name, func(t *testing.T) {
			analyzer := initcheck.New(testCase.cfg)
			analysistest.Run(t, testdata, analyzer, testCase.pkg...)
		})
	}
}

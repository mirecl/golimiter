package genericcheck_test

import (
	"testing"

	"github.com/mirecl/golimiter/pkg/genericcheck"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestGeneric(t *testing.T) {
	testdata := analysistest.TestData()

	TestCases := []struct {
		name string
		pkg  []string
	}{
		{
			name: "forbided type/func",
			pkg:  []string{"a"},
		},
		{
			name: "forbided type #2",
			pkg:  []string{"b"},
		},
		{
			name: "forbided type #3",
			pkg:  []string{"c"},
		},
		{
			name: "forbided type #4",
			pkg:  []string{"d"},
		},
		{
			name: "success analysis",
			pkg:  []string{"e"},
		},
	}

	for _, testCase := range TestCases {
		t.Run(testCase.name, func(t *testing.T) {
			analyzer := genericcheck.New()
			analysistest.Run(t, testdata, analyzer, testCase.pkg...)
		})
	}
}

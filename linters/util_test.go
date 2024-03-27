package linters

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetSegments(t *testing.T) {
	tests := []struct {
		ident        string
		wantSegments []string
	}{
		{"", nil},
		{"a", []string{"a"}},
		{"aB", []string{"a", "B"}},
		{"a123b", []string{"a123b"}},
		{"a123b4", []string{"a123b4"}},
		{"a123B", []string{"a123", "B"}},
		{"A123B", []string{"A123", "B"}},
		{"aBBBCd", []string{"a", "BBB", "Cd"}},
		{"ABCDa", []string{"ABC", "Da"}},
		{"xABCDa", []string{"x", "ABC", "Da"}},
		{"abc12a34", []string{"abc12a34"}},
		{"abc12A34", []string{"abc12", "A34"}},
		{"aBc12A34", []string{"a", "Bc12", "A34"}},
	}
	for _, tt := range tests {
		t.Run(tt.ident, func(t *testing.T) {
			require.Equal(t, tt.wantSegments, GetSegments(tt.ident))
		})
	}
}

func TestFindIdentsWithCommonPrefix(t *testing.T) {

	tests := []struct {
		prefixSource     string
		idents           []string
		wantCommonPrefix string
		wantFound        []string
	}{
		{"RedFox", []string{"redFox1", "Red2", "Field1"}, "Red", []string{"redFox1", "Red2"}},
		{"RedFox", []string{"redFox1", "RedFox2", "Field1"}, "RedFox", []string{"redFox1", "RedFox2"}},
		{"Prefix", []string{"redFox1", "RedFox2", "Field1"}, "", nil},
	}
	for i, tt := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			gotCommonPrefix, gotFound := FindIdentsWithPartialPrefix(tt.prefixSource, tt.idents)
			require.Equal(t, tt.wantCommonPrefix, gotCommonPrefix)
			require.Equal(t, tt.wantFound, gotFound)
		})
	}
}

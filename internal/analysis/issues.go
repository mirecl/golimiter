package analysis

import (
	"crypto/md5"
	"encoding/hex"
	"go/token"
	"os"
	"path/filepath"
)

// Issue problem in analysis.
type Issue struct {
	Message  string         `json:"message"`
	Code     string         `json:"code"`
	Pos      token.Pos      `json:"omitempty"`
	End      token.Pos      `json:"omitempty"`
	Fset     *token.FileSet `json:"omitempty"`
	Filename string         `json:"filename"`
	Line     int            `json:"line"`
	Hash     string         `json:"hash"`
}

func (i Issue) Position() string {
	position := i.Fset.Position(i.Pos).String()

	dir, err := os.Getwd()
	if err != nil {
		return position
	}

	p, err := filepath.Rel(dir, position)
	if err != nil {
		return position
	}

	return p
}

func GetHash(fset *token.FileSet, start, end token.Pos) string {
	b, err := os.ReadFile(fset.Position(start).Filename)
	if err != nil {
		return "Unknown"
	}

	hash := md5.Sum(b[fset.Position(start).Offset:fset.Position(end).Offset])
	return hex.EncodeToString(hash[:])
}

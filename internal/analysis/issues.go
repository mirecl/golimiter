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

func (i Issue) Hash() string {
	b, err := os.ReadFile(i.Fset.Position(i.Pos).Filename)
	if err != nil {
		return "Unknown"
	}

	start := i.Fset.Position(i.Pos).Offset
	end := i.Fset.Position(i.End).Offset

	hash := md5.Sum(b[start:end])
	return hex.EncodeToString(hash[:])
}

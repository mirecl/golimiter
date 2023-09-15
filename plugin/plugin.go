package main

import (
	"encoding/json"
	"fmt"

	"github.com/mirecl/golimiter/internal/config"
	"github.com/mirecl/golimiter/pkg/exprcheck"
	"github.com/mirecl/golimiter/pkg/genericcheck"
	"github.com/mirecl/golimiter/pkg/goroutinecheck"
	"github.com/mirecl/golimiter/pkg/initcheck"
	"golang.org/x/tools/go/analysis"
)

// New plugin for golangci-lint.
func New(conf any) ([]*analysis.Analyzer, error) {
	b, err := json.Marshal(conf)
	if err != nil {
		return nil, fmt.Errorf("failed read `config`: %w", err)
	}

	cfg := config.Default()
	if err := json.Unmarshal(b, &cfg); err != nil {
		return nil, fmt.Errorf("failed convert `config`: %w", err)
	}

	return []*analysis.Analyzer{
		initcheck.New(&cfg.Init),
		goroutinecheck.New(&cfg.Goroutine),
		exprcheck.New(&cfg.Expr),
		genericcheck.New(),
	}, nil
}

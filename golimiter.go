package main

import (
	"github.com/mirecl/golimiter/internal/config"
	"github.com/mirecl/golimiter/pkg/exprcheck"
	"github.com/mirecl/golimiter/pkg/goroutinecheck"
	"github.com/mirecl/golimiter/pkg/initcheck"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	cfg := config.Read()

	multichecker.Main(
		initcheck.New(&cfg.Init),
		goroutinecheck.New(&cfg.Goroutine),
		exprcheck.New(&cfg.Expr),
	)
}

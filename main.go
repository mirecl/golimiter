package main

import (
	"github.com/mirecl/golimiter/internal/analysis"
	"github.com/mirecl/golimiter/internal/config"
	"github.com/mirecl/golimiter/internal/linters"
)

func init() {
	config.Read()
}

func main() {
	analysis.Run(
		linters.NewNoGeneric(),
		linters.NewNoInit(),
		linters.NewNoGoroutine(),
	)
}

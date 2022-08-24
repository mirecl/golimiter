package main

import (
	"github.com/mirecl/golimiter/pkg/goroutinecheck"
	"github.com/mirecl/golimiter/pkg/initcheck"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	// TODO: set config.
	Limit := 0
	multichecker.Main(
		goroutinecheck.New(new(goroutinecheck.Config)),
		initcheck.New(&initcheck.Config{Limit: &Limit}),
	)
}

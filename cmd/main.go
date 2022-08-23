package main

import (
	"github.com/mirecl/golimiter/pkg/goroutinecheck"
	"github.com/mirecl/golimiter/pkg/initcheck"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	Allowed := 10
	multichecker.Main(
		goroutinecheck.New(new(goroutinecheck.Config)),
		initcheck.New(&initcheck.Config{Allowed: &Allowed}),
	)
}

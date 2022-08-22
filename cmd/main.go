package main

import (
	"github.com/mirecl/golimiter/pkg/goroutinecheck"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		goroutinecheck.New(new(goroutinecheck.Config)),
	)
}

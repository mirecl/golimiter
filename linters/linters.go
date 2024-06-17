package linters

import "github.com/mirecl/golimiter/analysis"

// All lintres for analysis.
var All = []*analysis.Linter{
	NewNoGeneric(),
	NewNoInit(),
	NewNoGoroutine(),
	NewNoNoLint(),
	NewNoDefer(),
	NewNoLength(),
	NewNoPrefix(),
	NewNoUnderscore(),
	NewNoObject(),
}

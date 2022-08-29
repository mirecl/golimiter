package c

import (
	"a"
	"b"
)

type T[P b.Constraint] a.T[P] // want "a `generic` statement forbidden to use."

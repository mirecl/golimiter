package a

type T[P any] int // want "a `generic` statement forbidden to use."

func F[P any]() {} // want "a `generic` statement forbidden to use."

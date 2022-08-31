package b

func main() {
	go func() {}() // want "a number of allowed `goroutine` statement 2."
	go func() {}() // want "a number of allowed `goroutine` statement 2."
	go func() {}() // want "a number of allowed `goroutine` statement 2."
}

package f

func main2() {
	go func() {}() // want "a `goroutine` statement forbidden to use."
}

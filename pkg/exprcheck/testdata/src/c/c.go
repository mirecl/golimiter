package c

func c1(a, b, c int) bool {
	return (a == b) || (a == c) // want "a complexity expr 3, allowed 0."
}

func c2(a, b, c int) bool {
	return (a == b) && (a == c) // want "a complexity expr 3, allowed 0."
}

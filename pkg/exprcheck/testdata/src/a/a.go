package a

func a(a, b, c int) bool {
	if (a == b && b != c) || a > c && (a == b) { // want "a complexity expr 7, allowed 0."
		return c != b // want "a complexity expr 1, allowed 0."
	} else if (a != b) || a > c { // want "a complexity expr 3, allowed 0."
		return true
	}
	return (a == b) == ((a > b) && (c != a)) // want "a complexity expr 5, allowed 0."
}

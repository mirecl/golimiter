package e

func a(a, b, c int) bool {
	if (a == b && b != c) || a > c && (a == b) { // want "a complexity expr 7, allowed 3."
		return c != b
	} else if (a != b) || a > c {
		return true
	}
	return (a == b) == ((a > b) && (c != a)) // want "a complexity expr 5, allowed 3."
}

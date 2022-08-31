package h

func h2(a, b, c int) bool {
	if (a == b && b != c) || a > c && (a == b) {
		return c != b
	} else if (a != b) || a > c {
		return true
	}
	return (a == b) == ((a > b) && (c != a))
}

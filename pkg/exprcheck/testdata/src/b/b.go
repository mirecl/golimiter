package b

func b1(a, b, c int) bool {
	if (a == b) || (a == c) { // want "a complexity expr 3, allowed 0."
		return true
	}
	return false
}

func b2(a, b, c int) bool {
	if (a == b) && (a == c) { // want "a complexity expr 3, allowed 0."
		return true
	}
	return false
}

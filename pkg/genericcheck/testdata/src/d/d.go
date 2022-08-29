package d

type myInt int

func _() {
	type constraint interface { // want "a `generic` statement forbidden to use."
		myInt
	}
}

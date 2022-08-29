package b

type Constraint interface { // want "a `generic` statement forbidden to use."
	~int | string
}

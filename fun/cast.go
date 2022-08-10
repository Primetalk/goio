package fun

type Any interface{}

// CastAsInterface casts a value of an arbitrary type as interface {}.
func CastAsInterface[A any](a A) Any {
	return a
}

// Cast converts interface {} to ordinary type A.
// It'a simple operation i.(A) represented as a function.
// In case the conversion is not possible, returns an error.
func Cast[A any](i Any) (a A, err error) {
	defer RecoverToErrorVar("Cast", &err)
	a = i.(A)
	return
}

// UnsafeCast converts interface {} to ordinary type A.
// It'a simple operation i.(A) represented as a function.
// In case the conversion is not possible throws a panic.
func UnsafeCast[A any](i Any) A {
	return i.(A)
}

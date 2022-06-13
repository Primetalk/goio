package fun

// CastAsInterface casts a value of an arbitrary type as interface {}.
func CastAsInterface[A any](a A) interface{} {
	return a
}

// UnsafeCast converts interface {} to ordinary type A.
// It'a simple operation i.(A) represented as a function.
// In case the conversion is not possible throws a panic.
func UnsafeCast[A any](i interface{}) A {
	return i.(A)
}

package fun

// Predicate is a function with a boolean result type.
type Predicate[A any] func(A) bool

// IsEqualTo compares two arguments for equality.
func IsEqualTo[A comparable](a A) Predicate[A] {
	return func(other A) bool {
		return a == other
	}
}

// Not negates the given predicate.
func Not[A any](p Predicate[A]) Predicate[A] {
	return func(a A) bool {
		return !p(a)
	}
}

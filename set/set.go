package set

// Set is a map with a dummy value.
type Set[A comparable] map[A]struct{}

// Contains creates a predicate that will check if an element is in this set.
func Contains[A comparable](set Set[A]) func(A) bool {
	return func(a A) (ok bool) {
		_, ok = set[a]
		return
	}
}

// SetSize returns the size of the set.
func SetSize[A comparable](s Set[A]) int {
	return len(s)
}

// ToSlice retrieves all set elements in an unpredictable order.
func (s Set[A]) ToSlice() (res []A) {
	res = make([]A, 0, len(s))
	for k := range s {
		res = append(res, k)
	}
	return
}

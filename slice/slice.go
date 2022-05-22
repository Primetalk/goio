// Package slice provides common utility functions to Go slices.
package slice

// Map converts all values of a slice using the provided function.
func Map[A any, B any](as []A, f func(A) B) (bs []B) {
	bs = make([]B, 0, len(as))
	for _, a := range as {
		bs = append(bs, f(a))
	}
	return
}

// FlatMap converts all values of a slice using the provided function.
// As the function returns slices, all of them are appended to a single long slice.
func FlatMap[A any, B any](as []A, f func(A) []B) (bs []B) {
	bs = make([]B, 0, len(as))
	for _, a := range as {
		bs = append(bs, f(a)...)
	}
	return
}

// FoldLeft folds all values in the slice using the combination function.
func FoldLeft[A any, B any](as []A, zero B, f func(B, A) B) (res B) {
	res = zero
	for _, a := range as {
		res = f(res, a)
	}
	return
}

// Filter filters slice values.
func Filter[A any](as []A, p func(a A) bool) (res []A) {
	res = make([]A, 0, len(as))
	for _, a := range as {
		if p(a) {
			res = append(res, a)
		}
	}
	return
}

// FilterNot filters slice values inverting the condition.
func FilterNot[A any](as []A, p func(a A) bool) (res []A) {
	res = make([]A, 0, len(as))
	for _, a := range as {
		if !p(a) {
			res = append(res, a)
		}
	}
	return
}

// Flatten simplifies a slice of slices to just a slice.
func Flatten[A any](ass [][]A) (aas []A) {
	total := 0
	for _, as := range ass {
		total += len(as)
	}
	aas = make([]A, 0, total)
	for _, as := range ass {
		aas = append(aas, as...)
	}
	return
}

// Set is a way to represent sets in Go.
type Set[A comparable] map[A]struct{}

// ToSet converts a slice to a set.
func ToSet[A comparable](as []A) (s Set[A]) {
	s = make(map[A]struct{}, len(as))
	for _, a := range as {
		s[a] = struct{}{}
	}
	return
}

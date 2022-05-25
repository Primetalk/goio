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

// AppendAll concatenates all slices.
func AppendAll[A any](ass ...[]A) (aas []A) {
	return Flatten(ass)
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

// SetSize returns the size of the set.
func SetSize[A comparable](s Set[A]) int {
	return len(s)
}

// GroupBy groups elements by a function that returns a key.
func GroupBy[A any, K comparable](as []A, f func(A) K) (res map[K][]A) {
	res = make(map[K][]A, len(as))
	for _, a := range as {
		k := f(a)
		sl, ok := res[k]
		if ok {
			sl = append(sl, a)
			res[k] = sl
		} else {
			res[k] = []A{a}
		}
	}
	return
}

// Sliding splits the provided slice into windows.
// Each window will have the given size.
// The first window starts from offset = 0.
// Each consequtive window starts at prev_offset + step.
// Last window might very well be shorter.
func Sliding[A any](as []A, size int, step int) (res [][]A) {
	for offset := 0; offset < len(as); offset += step {
		high := offset + size
		if high > len(as) {
			high = len(as)
		}
		slice1 := as[offset:high]
		res = append(res, slice1)
		if high == len(as) {
			break
		}
	}
	return
}

// Grouped partitions the slice into groups of the given size.
// Last partition might be smaller.
func Grouped[A any](as []A, size int) (res [][]A) {
	return Sliding(as, size, size)
}

// Package slice provides common utility functions to Go slices.
package slice

import (
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/maps"
	"github.com/primetalk/goio/option"
	"github.com/primetalk/goio/set"
)

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

// Reduce aggregates all elements pairwise.
// Only works for non empty slices.
func Reduce[A any](as []A, f func(A, A) A) A {
	return FoldLeft(as[1:], as[0], f)
}

// Filter filters slice values.
func Filter[A any](as []A, p fun.Predicate[A]) (res []A) {
	res = make([]A, 0, len(as))
	for _, a := range as {
		if p(a) {
			res = append(res, a)
		}
	}
	return
}

// FilterNot filters slice values inverting the condition.
func FilterNot[A any](as []A, p fun.Predicate[A]) (res []A) {
	res = make([]A, 0, len(as))
	for _, a := range as {
		if !p(a) {
			res = append(res, a)
		}
	}
	return
}

// Partition separates elements in as according to the predicate.
func Partition[A any](as []A, p fun.Predicate[A]) (resT []A, resF []A) {
	resT = make([]A, 0, len(as))
	resF = make([]A, 0, len(as))
	for _, a := range as {
		if p(a) {
			resT = append(resT, a)
		} else {
			resF = append(resF, a)
		}
	}
	return
}

// Count counts the number of elements that satisfy the given predicate.
func Count[A any](as []A, predicate fun.Predicate[A]) (cnt int) {
	cnt = 0
	for _, a := range as {
		if predicate(a) {
			cnt += 1
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

// ToSet converts a slice to a set.
func ToSet[A comparable](as []A) (s set.Set[A]) {
	s = make(map[A]struct{}, len(as))
	for _, a := range as {
		s[a] = struct{}{}
	}
	return
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

// GroupByMap is a convenience function that groups and then maps the subslices.
func GroupByMap[A any, K comparable, B any](as []A, f func(A) K, g func([]A) B) (res map[K]B) {
	intermediate := GroupBy(as, f)
	return maps.MapValues(intermediate, g)
}

// GroupByMapCount for each key counts how often it is seen.
func GroupByMapCount[A any, K comparable](as []A, f func(A) K) (res map[K]int) {
	return GroupByMap(as, f, Len[A])
}

// Sliding splits the provided slice into windows.
// Each window will have the given size.
// The first window starts from offset = 0.
// Each consecutive window starts at prev_offset + step.
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

// Len returns the length of the slice.
// This is a normal function that can be passed around unlike the built-in `len`.
func Len[A any](as []A) int {
	return len(as)
}

// Collect runs through the slice, executes the given function and
// only keeps good returned values.
func Collect[A any, B any](as []A, f func(a A) option.Option[B]) (bs []B) {
	for _, a := range as {
		bo := f(a)
		option.ForEach(bo, func(b B) {
			bs = append(bs, b)
		})
	}
	return
}

// Exists returns a predicate on slices.
// The predicate is true if there is an element that satisfy the given element-wise predicate.
// It's false for an empty slice.
func Exists[A any](p fun.Predicate[A]) fun.Predicate[[]A] {
	return func(as []A) (res bool) {
		res = false
		for _, a := range as {
			if p(a) {
				res = true
				break
			}
		}
		return
	}
}

// Forall returns a predicate on slices.
// The predicate is true if all elements satisfy the given element-wise predicate.
// It's true for an empty slice.
func Forall[A any](p fun.Predicate[A]) fun.Predicate[[]A] {
	return func(as []A) (res bool) {
		res = true
		for _, a := range as {
			if !p(a) {
				res = false
				break
			}
		}
		return
	}
}

// ForEach executes the given function for each element of the slice.
func ForEach[A any](as []A, f func(a A)) {
	for _, a := range as {
		f(a)
	}
}

// ZipWith returns a slice of pairs made of elements of the two slices.
// The length of the result is min of both.
func ZipWith[A any, B any](as []A, bs []B) (res []fun.Pair[A, B]) {
	l := len(as)
	lb := len(bs)
	if lb < l {
		l = lb
	}
	res = make([]fun.Pair[A, B], l)
	for i := 0; i < l; i++ {
		res[i] = fun.NewPair(as[i], bs[i])
	}
	return
}

// ZipWithIndex prepends the index to each element.
func ZipWithIndex[A any](as []A) (res []fun.Pair[int, A]) {
	res = make([]fun.Pair[int, A], len(as))
	for i, a := range as {
		res[i] = fun.NewPair(i, a)
	}
	return
}

// IndexOf returns the index of the first occurrence of a in the slice
// or -1 if not found.
func IndexOf[A comparable](as []A, a A) int {
	for i, a1 := range as {
		if a == a1 {
			return i
		}
	}
	return -1
}

// Take returns at most n elements.
func Take[A any](as []A, n int) []A {
	return as[:fun.Min(n, len(as))]
}

// Drop removes initial n elements.
func Drop[A any](as []A, n int) []A {
	return as[fun.Min(n, len(as)):]
}

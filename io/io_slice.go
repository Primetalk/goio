package io

import "github.com/primetalk/goio/slice"

// MapSlice converts each element of the slice inside IO[[]A] using the provided function that cannot fail.
func MapSlice[A any, B any](ioas IO[[]A], f func(a A) B) IO[[]B] {
	return Map(ioas, func(as []A) []B { return slice.Map(as, f) })
}

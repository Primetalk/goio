package io

import (
	"github.com/primetalk/goio/fun"
)

// Memoize returns a function that will remember the original function in a map.
// It's thread safe, however, not super performant.
func Memoize[A comparable, B any](f func(a A) IO[B]) func(A) IO[B] {
	m := fun.Memoize(IOFuncToGoResult(f))
	return func(a A) IO[B] {
		return Delay(func() IO[B] {
			return FromConstantGoResult(m(a))
		})
	}
}

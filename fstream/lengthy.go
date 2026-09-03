package fstream

import (
	"github.com/primetalk/goio/io"
)

// Repeat appends the same stream infinitely.
func Repeat[A any](stm Stream[A]) Stream[A] {
	return AndThenLazy(stm, IOStream[A](io.Pure(func() Stream[A] { return Repeat(stm) })))
}

// Take cuts the stream after n elements.
func Take[A any](stm Stream[A], n int) IOStream[A] {
	if n <= 0 {
		return IOStream[A](EmptyIO[A]())
	} else {
		return IOStream[A](StreamMatch(stm,
			func() io.IO[Stream[A]] { return io.IO[Stream[A]](EmptyIO[A]()) },
			func(a A, tail Stream[A]) io.IO[Stream[A]] {
				return io.IO[Stream[A]](AndThenLazy(Lift(a), Take(tail, n-1)).ToIOStream())
			},
			func(tail Stream[A]) io.IO[Stream[A]] {
				return io.IO[Stream[A]](Take(tail, n))
			},
			func(err error) io.IO[Stream[A]] {
				return io.IO[Stream[A]](io.Fail[Stream[A]](err))
			},
		))
	}
}

// Drop skips n elements in the stream.
func Drop[A any](stm Stream[A], n int) IOStream[A] {
	if n <= 0 {
		return stm.ToIOStream()
	} else {
		return IOStream[A](StreamMatch(stm,
			func() io.IO[Stream[A]] { return io.IO[Stream[A]](EmptyIO[A]()) },
			func(a A, tail Stream[A]) io.IO[Stream[A]] {
				return io.IO[Stream[A]](Drop(tail, n-1))
			},
			func(tail Stream[A]) io.IO[Stream[A]] {
				return io.IO[Stream[A]](Drop(tail, n))
			},
			func(err error) io.IO[Stream[A]] {
				return io.IO[Stream[A]](io.Fail[Stream[A]](err))
			},
		))
	}
}

// TakeWhile returns the beginning of the stream such that all elements satisfy the predicate.
func TakeWhile[A any](stm Stream[A], predicate func(A) bool) IOStream[A] {
	return IOStreamMatch(stm,
		EmptyIO[A],
		func(a A, tail Stream[A]) IOStream[A] {
			if predicate(a) {
				return AndThenLazy(Lift(a), TakeWhile(tail, predicate)).ToIOStream()
			} else {
				return EmptyIO[A]()
			}
		},
		func(tail Stream[A]) IOStream[A] {
			return TakeWhile(tail, predicate)
		},
		func(err error) IOStream[A] {
			return IOStream[A](io.Fail[Stream[A]](err))
		},
	)
}

// DropWhile removes the beginning of the stream so that the new stream starts with an element
// that falsifies the predicate.
func DropWhile[A any](stm Stream[A], predicate func(A) bool) IOStream[A] {
	return IOStreamMatch(stm,
		EmptyIO[A],
		func(a A, tail Stream[A]) IOStream[A] {
			if predicate(a) {
				return DropWhile(tail, predicate)
			} else {
				return AndThen(Lift(a), tail).ToIOStream()
			}
		},
		func(tail Stream[A]) IOStream[A] {
			return DropWhile(tail, predicate)
		},
		func(err error) IOStream[A] {
			return IOStream[A](io.Fail[Stream[A]](err))
		},
	)
}

package stream

import (
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
)

// ChunkN groups elements by n and produces a stream of slices.
func ChunkN[A any](n int) func(sa Stream[A]) Stream[[]A] {
	return ToChunks[A](n)
}

// ToChunks collects incoming elements in chunks of the given size.
func ToChunks[A any](size int) func(stm Stream[A]) Stream[[]A] {
	return func(stm Stream[A]) Stream[[]A] {
		return StateFlatMapWithFinish(stm, make([]A, 0, size),
			func(a A, as []A) io.IO[fun.Pair[[]A, Stream[[]A]]] {
				return io.Pure(func() fun.Pair[[]A, Stream[[]A]] {
					as2 := append(as, a)
					if len(as2) >= size {
						return fun.NewPair(make([]A, 0, size), Lift(as2))
					} else {
						return fun.NewPair(as2, Empty[[]A]())
					}
				})
			},
			func(as []A) Stream[[]A] {
				return Lift(as)
			},
		)
	}
}

// ChunksResize rebuffers chunks to the given size.
func ChunksResize[A any](newSize int) func(stm Stream[[]A]) Stream[[]A] {
	return func(stm Stream[[]A]) Stream[[]A] {
		return StateFlatMapWithFinish(stm, []A{},
			func(as1 []A, st []A) io.IO[fun.Pair[[]A, Stream[[]A]]] {
				return io.Pure(func() fun.Pair[[]A, Stream[[]A]] {
					st2 := append(st, as1...)
					cnt := len(st2) / newSize
					chunks := [][]A{}
					for i := 0; i < cnt; i++ {
						chunks = append(chunks, st2[i*newSize:(i+1)*newSize])
					}
					last := st2[cnt*newSize:]
					return fun.NewPair(last, LiftMany(chunks...))
				})
			},
			func(st []A) Stream[[]A] {
				return Lift(st)
			},
		)
	}
}

package stream

import (
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
)

// ZipWithIndex prepends the index to each element.
func ZipWithIndex[A any](as Stream[A]) Stream[fun.Pair[int, A]] {
	return StateFlatMap(as, 0,
		func(a A, i int) io.IO[fun.Pair[int, Stream[fun.Pair[int, A]]]] {
			return io.Lift(
				fun.NewPair(i+1, Lift(
					fun.NewPair(i, a),
				)),
			)
		},
	)
}

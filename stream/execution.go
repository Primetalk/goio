package stream

import (
	"errors"
	"fmt"

	"github.com/primetalk/goio/either"
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/option"
)

// Collector reads the stream and produces some value.
type Collector[A any, B any] func(Stream[A]) io.IO[B]

// Collect collects all element from the stream and for each element invokes
// the provided function
func Collect[A any](stm Stream[A], collector func(A) error) io.IO[fun.Unit] {
	return io.FlatMap(
		io.IO[StepResult[A]](stm),
		func(sra StepResult[A]) io.IO[fun.Unit] {
			if sra.IsFinished {
				return io.Lift(fun.Unit1)
			} else {
				rest := Collect(sra.Continuation, collector)
				if sra.HasValue {
					return io.AndThen(io.FromUnit(func() error {
						return collector(sra.Value)
					}), rest)
				} else {
					return rest
				}
			}
		})
}

// ForEach invokes a simple function for each element of the stream.
func ForEach[A any](stm Stream[A], collector func(A)) io.IO[fun.Unit] {
	return Collect(stm, func(a A) error {
		collector(a)
		return nil
	})
}

// DrainAll executes the stream and throws away all values.
func DrainAll[A any](stm Stream[A]) io.IO[fun.Unit] {
	return io.FlatMap(
		io.IO[StepResult[A]](stm),
		func(sra StepResult[A]) io.IO[fun.Unit] {
			if sra.IsFinished {
				return io.Lift(fun.Unit1)
			} else {
				return DrainAll(sra.Continuation)
			}
		})
}

// AppendToSlice executes the stream and appends it's results to the slice.
func AppendToSlice[A any](stm Stream[A], start []A) io.IO[[]A] {
	return io.FlatMap(
		io.IO[StepResult[A]](stm),
		func(sra StepResult[A]) io.IO[[]A] {
			if sra.IsFinished {
				return io.Lift(start)
			} else if sra.HasValue {
				return AppendToSlice(sra.Continuation, append(start, sra.Value))
			} else {
				return AppendToSlice(sra.Continuation, start)
			}
		})
}

// ToSlice executes the stream and collects all results to a slice.
func ToSlice[A any](stm Stream[A]) io.IO[[]A] {
	return AppendToSlice(stm, []A{})
}

// ErrHeadOfEmptyStream - an error that is returned when someone attempts to retrieve
// the head of an empty stream.
var ErrHeadOfEmptyStream = errors.New("head of empty stream")

// HeadAndTail returns the very first element of the stream and the rest of the stream.
func HeadAndTail[A any](stm Stream[A]) io.IO[fun.Pair[A, Stream[A]]] {
	return StreamFold(stm,
		func() io.IO[fun.Pair[A, Stream[A]]] {
			return io.Fail[fun.Pair[A, Stream[A]]](ErrHeadOfEmptyStream)
		},
		func(a A, tail Stream[A]) io.IO[fun.Pair[A, Stream[A]]] {
			return io.Lift(fun.NewPair(a, tail))
		},
		func(tail Stream[A]) io.IO[fun.Pair[A, Stream[A]]] {
			return HeadAndTail(tail)
		},
		func(err error) io.IO[fun.Pair[A, Stream[A]]] {
			return io.Fail[fun.Pair[A, Stream[A]]](err)
		},
	)
}

// Head returns the first element of the stream.
// It'll fail if the stream is empty.
func Head[A any](stm Stream[A]) io.IO[A] {
	slice1 := ToSlice(Take(stm, 1))
	return io.MapErr(slice1, func(as []A) (a A, err error) {
		if len(as) > 0 {
			a = as[0]
		} else {
			err = ErrHeadOfEmptyStream
		}
		return
	})
}

// Last keeps track of the current element of the stream
// and returns it when the stream completes.
func Last[A any](stm Stream[A]) io.IO[A] {
	optA := FoldLeft(
		stm, option.None[A](),
		func(st option.Option[A], a A) option.Option[A] {
			return option.Some(a) // ignore the previous last value
		},
	)
	return io.FlatMap(
		optA,
		func(st option.Option[A]) io.IO[A] {
			return option.Fold(st,
				io.Lift[A],
				func() io.IO[A] {
					return io.Fail[A](fmt.Errorf("last of an empty stream"))
				},
			)
		},
	)
}

// Partition divides the stream into two that are handled independently.
func Partition[A any, C any, D any](stm Stream[A],
	predicate func(A) bool,
	trueHandler Collector[A, C],
	falseHandler Collector[A, D],
) io.IO[fun.Pair[C, D]] {
	eithersIO := FanOut(stm,
		func(stm Stream[A]) io.IO[either.Either[C, D]] {
			return io.Map(trueHandler(Filter(stm, predicate)), either.Left[C, D])
		},
		func(stm Stream[A]) io.IO[either.Either[C, D]] {
			return io.Map(falseHandler(FilterNot(stm, predicate)), either.Right[C, D])
		},
	)
	return io.Map(eithersIO, func(eithers []either.Either[C, D]) (p fun.Pair[C, D]) {
		if either.IsLeft(eithers[0]) {
			p = fun.NewPair(eithers[0].Left, eithers[1].Right)
		} else {
			p = fun.NewPair(eithers[1].Left, eithers[0].Right)
		}
		return
	})
}

// TakeAndTail collects n leading elements of the stream and
// returns them along with the tail of the stream.
// If the stream is shorter, then only available elements are returned and an emtpy stream.
func TakeAndTail[A any](stm Stream[A], n int, prefix []A) io.IO[fun.Pair[[]A, Stream[A]]] {
	if n == 0 {
		return io.Lift(fun.NewPair(prefix, stm))
	} else {
		return StreamFold(stm,
			func() io.IO[fun.Pair[[]A, Stream[A]]] {
				return io.Lift(fun.NewPair(prefix, Empty[A]()))
			},
			func(a A, tail Stream[A]) io.IO[fun.Pair[[]A, Stream[A]]] {
				return TakeAndTail(tail, n-1, append(prefix, a))
			},
			func(tail Stream[A]) io.IO[fun.Pair[[]A, Stream[A]]] {
				return TakeAndTail(tail, n, prefix)
			},
			func(err error) io.IO[fun.Pair[[]A, Stream[A]]] {
				return io.Fail[fun.Pair[[]A, Stream[A]]](err)
			},
		)
	}
}

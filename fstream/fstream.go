package fstream

import (
	"log"

	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/option"
	"github.com/primetalk/goio/stream"
)

type Scope struct {
	Finalizer io.IOUnit
}

type ScopedStream[A any] struct {
	sa        stream.Stream[A]
	finalizer io.IOUnit
}

func NewScopedStream[A any](sa stream.Stream[A], finalizer io.IOUnit) ScopedStream[A] {
	return ScopedStream[A]{
		sa:        sa,
		finalizer: finalizer,
	}
}

func (stm ScopedStream[A]) Close() io.IOUnit {
	return stm.finalizer
}

func ScopedStreamMatch[A any, B any](stm ScopedStream[A],
	onFinish func() io.IO[B],
	onValue func(a A, tail ScopedStream[A]) io.IO[B],
	onEmpty func(tail ScopedStream[A]) io.IO[B],
	onError func(err error) io.IO[B],
) io.IO[B] {
	return stream.StreamMatch(stm.sa,
		func() io.IO[B] {
			return io.AndThen(stm.finalizer, onFinish())
		},
		func(a A, tail stream.Stream[A]) io.IO[B] {
			return onValue(a, NewScopedStream(tail, stm.finalizer))
		},
		func(tail stream.Stream[A]) io.IO[B] {
			return onEmpty(NewScopedStream(tail, stm.finalizer))
		},
		func(err error) io.IO[B] {
			return io.Fold(stm.finalizer,
				func(u fun.Unit) io.IO[B] {
					return onError(err)
				},
				func(err2 error) io.IO[B] {
					log.Printf("double error during fstream.ScopedStreamMatch: %+v", err2)
					return onError(err)
				})
		},
	)
}

// fstream.Stream is similar to stream.Stream, but it keeps track of
// a finalizer for the stream
// - an action that needs to be executed after the stream is finished.
// We maintain finalizer in all operations.
// Finalizer is executed regardless of the stream failure.
// After finalizer, some tail stream could be returned.
// Finalizer is the very first operation of the second stream.
type Stream[A any] struct {
	init ScopedStream[A]
	cont option.Option[IOStream[A]]
}

type NotYetOpenedStream[A any] io.IO[Stream[A]]
type IOStream[A any] io.IO[Stream[A]]

func IOStreamFlatMap[A any, B any](ios IOStream[A], f func(Stream[A]) io.IO[B]) io.IO[B] {
	return io.FlatMap(io.IO[Stream[A]](ios), f)
}

type Pipe[A any, B any] func(Stream[A]) Stream[B]
type IOPipe[A any, B any] func(Stream[A]) io.IO[Stream[B]]

// NewStreamWithConfinuation constructs a fresh instance of stream.
func NewStream[A any](init ScopedStream[A], last option.Option[IOStream[A]]) Stream[A] {
	return Stream[A]{
		init: init,
		cont: last,
	}
}

// NewStreamWithConfinuation constructs a fresh instance of stream.
func NewStreamWithConfinuation[A any](init stream.Stream[A], finalizer io.IOUnit, last IOStream[A]) Stream[A] {
	return NewStream(NewScopedStream(init, finalizer), option.Some(last))
}

// NewStreamWithoutConfinuation constructs a fresh instance of stream.
func NewStreamWithoutConfinuation[A any](init stream.Stream[A], finalizer io.IOUnit) Stream[A] {
	return NewStream(NewScopedStream(init, finalizer), option.None[IOStream[A]]())
}

func EmptyIO[A any]() IOStream[A] {
	return IOStream[A](io.Delay(func() io.IO[Stream[A]] { return io.Lift(Empty[A]()) }))
}

func Empty[A any]() Stream[A] {
	return NewStreamWithoutConfinuation(stream.Empty[A](), io.IOUnit1)
}

// Lift converts ordinary stream to fstream.Stream.
func Lift[A any](a A) Stream[A] {
	return LiftStream(stream.Lift(a))
}

// LiftIO returns an io-stream of a single element a.
func LiftIO[A any](a A) IOStream[A] {
	return Lift(a).ToIOStream()
}

// Lift converts ordinary stream to fstream.Stream.
func LiftMany[A any](as ...A) Stream[A] {
	return LiftStream(stream.LiftMany(as...))
}

// Lift converts ordinary stream to fstream.Stream.
func LiftStream[A any](sa stream.Stream[A]) Stream[A] {
	return NewStreamWithoutConfinuation(sa, io.IOUnit1)
}

// LiftStreamIO converts ordinary stream to fstream.Stream.
func LiftStreamIO[A any](sa stream.Stream[A]) IOStream[A] {
	return IOStream[A](io.Delay(func() io.IO[Stream[A]] { return io.Lift(LiftStream(sa)) }))
}

func (sa Stream[A]) ToIOStream() IOStream[A] {
	return IOStream[A](io.Lift(sa))
}

func (iosa IOStream[A]) ToStream() Stream[A] {
	return NewStream(NewScopedStream(stream.Empty[A](), io.IOUnit1), option.Some(iosa))
}

// ConcatPlainStreams concanenates a pair of ordinary streams.
func ConcatPlainStreams[A any](sa1 stream.Stream[A], sa2 stream.Stream[A]) (res Stream[A]) {
	return NewStreamWithConfinuation(sa1, io.IOUnit1, LiftStreamIO(sa2))
}

// StreamMatch performs arbitrary processing of a stream's single step result.
// in case of errors, it'll first run finalizer and then the user function.
func StreamMatch[A any, B any](
	stm Stream[A],
	onFinish func() io.IO[B],
	onValue func(a A, tail Stream[A]) io.IO[B],
	onEmpty func(tail Stream[A]) io.IO[B],
	onError func(err error) io.IO[B],
) io.IO[B] {
	return stream.StreamMatch(stm.init.sa,
		func() io.IO[B] {
			afterFinalizer := option.Match(stm.cont,
				func(cont IOStream[A]) io.IO[B] {
					return io.FlatMap(io.IO[Stream[A]](cont), func(stm2 Stream[A]) io.IO[B] {
						return StreamMatch(stm2, onFinish, onValue, onEmpty, onError)
					})
				},
				onFinish,
			)
			return io.AndThen(stm.init.finalizer, afterFinalizer)
		},
		func(a A, tail stream.Stream[A]) io.IO[B] {
			return onValue(a, NewStream(NewScopedStream(tail, stm.init.finalizer), stm.cont))
		},
		func(tail stream.Stream[A]) io.IO[B] {
			return onEmpty(NewStream(NewScopedStream(tail, stm.init.finalizer), stm.cont))
		},
		func(err error) io.IO[B] {
			// from the finalizer stream we only need the first side effect.
			return io.AndThen(stm.init.finalizer, onError(err))
		},
	)
}

func IOStreamMatch[A any, B any](
	stm Stream[A],
	onFinish func() IOStream[B],
	onValue func(a A, tail Stream[A]) IOStream[B],
	onEmpty func(tail Stream[A]) IOStream[B],
	onError func(err error) IOStream[B],
) IOStream[B] {
	return IOStream[B](StreamMatch(stm,
		func() io.IO[Stream[B]] { return io.IO[Stream[B]](onFinish()) },
		func(a A, tail Stream[A]) io.IO[Stream[B]] { return io.IO[Stream[B]](onValue(a, tail)) },
		func(tail Stream[A]) io.IO[Stream[B]] { return io.IO[Stream[B]](onEmpty(tail)) },
		func(err error) io.IO[Stream[B]] { return io.IO[Stream[B]](onError(err)) },
	))
}

// MapEval maps the values of the stream. The provided function returns an IO.
func MapEval[A any, B any](stm Stream[A], f func(a A) io.IO[B]) Stream[B] {
	return NewStream(NewScopedStream(stream.MapEval(stm.init.sa, f), stm.init.finalizer),
		option.Option[IOStream[B]](option.Map(stm.cont, func(iosa IOStream[A]) IOStream[B] {
			return IOStream[B](io.Map(io.IO[Stream[A]](iosa), func(sa Stream[A]) Stream[B] {
				return MapEval(sa, f)
			}))
		})),
	)
}

// Map converts values of the stream.
func Map[A any, B any](stm Stream[A], f func(a A) B) Stream[B] {
	return NewStream(NewScopedStream(stream.Map(stm.init.sa, f), stm.init.finalizer),
		option.Map(stm.cont, func(iosa IOStream[A]) IOStream[B] {
			return IOStream[B](io.Map(io.IO[Stream[A]](iosa), func(sa Stream[A]) Stream[B] {
				return Map(sa, f)
			}))
		}),
	)
}

// MapPipe creates a pipe that maps one stream through the provided function.
func MapPipe[A any, B any](f func(a A) B) Pipe[A, B] {
	return func(sa Stream[A]) Stream[B] {
		return Map(sa, f)
	}
}

// AndThen appends another stream after the end of the first one.
// Deprecated. The second stream won't be finalized.
func AndThen[A any](stm1 Stream[A], stm2 Stream[A]) Stream[A] {
	return AndThenLazy(stm1, IOStream[A](io.Lift(stm2)))
}

// AndThenLazy appends another stream. The other stream is constructed lazily.
// If the second stream has not constructed, it won't be finalized as well.
func AndThenLazy[A any](stm1 Stream[A], stm2 IOStream[A]) Stream[A] {
	return option.Match(stm1.cont,
		func(iosa IOStream[A]) Stream[A] {
			return NewStream(stm1.init, option.Some(
				IOStream[A](io.Map(io.IO[Stream[A]](iosa), func(sa Stream[A]) Stream[A] {
					return AndThenLazy(sa, stm2)
				})),
			))
		},
		func() Stream[A] {
			return NewStream(stm1.init, option.Some(stm2))
		},
	)
}

// IOAndThen concatenates a pair of io streams.
func IOAndThen[A any](stm1 IOStream[A], stm2 IOStream[A]) IOStream[A] {
	return IOStream[A](
		io.FlatMap(
			io.IO[Stream[A]](stm1),
			func(stm Stream[A]) io.IO[Stream[A]] {
				return io.Lift(AndThenLazy(stm, stm2))
			}),
	)
}

// FlatMap constructs a stream of streams.
func FlatMap[A any, B any](stm Stream[A], f func(a A) IOStream[B]) IOStream[B] {
	return IOStream[B](StreamMatch(stm,
		func() io.IO[Stream[B]] {
			return io.IO[Stream[B]](EmptyIO[B]())
		},
		func(a A, tail Stream[A]) io.IO[Stream[B]] {
			return io.Map(io.IO[Stream[B]](f(a)), func(sb Stream[B]) Stream[B] {
				return AndThenLazy(sb,
					IOStream[B](io.Delay(func() io.IO[Stream[B]] {
						return io.IO[Stream[B]](FlatMap(tail, f))
					})),
				)
			})
		},
		func(tail Stream[A]) io.IO[Stream[B]] {
			return io.IO[Stream[B]](FlatMap(tail, f))
		},
		func(err error) io.IO[Stream[B]] {
			return io.Fail[Stream[B]](err)
		},
	))
}

// FlatMapPipe creates a pipe that flatmaps one stream through the provided function.
func FlatMapPipe[A any, B any](f func(a A) IOStream[B]) IOPipe[A, B] {
	return func(sa Stream[A]) io.IO[Stream[B]] {
		return io.IO[Stream[B]](FlatMap(sa, f))
	}
}

// Flatten simplifies a stream of streams to just the stream of values by concatenating all
// inner streams.
func Flatten[A any](stm Stream[IOStream[A]]) IOStream[A] {
	return FlatMap(stm, fun.Identity[IOStream[A]])
}

// StateFlatMap maintains state along the way.
func StateFlatMap[A any, B any, S any](stm Stream[A], zero S, f func(a A, s S) (S, IOStream[B])) IOStream[B] {
	esb := EmptyIO[B]()
	onFinish := func(S) IOStream[B] { return esb }
	return StateFlatMapWithFinish(stm, zero, esb, f, onFinish)
}

// StateFlatMapWithFinish maintains state along the way.
// When the source stream finishes, it invokes onFinish with the last state.
func StateFlatMapWithFinish[A any, B any, S any](
	stm Stream[A],
	zero S,
	prefix IOStream[B],
	f func(a A, s S) (S, IOStream[B]),
	onFinish func(s S) IOStream[B],
) IOStream[B] {
	res := StreamMatch[A, Stream[B]](
		stm,
		/*onFinish*/ func() io.IO[Stream[B]] {
			return io.IO[Stream[B]](IOAndThen(prefix, onFinish(zero)))
		},
		/*onValue */ func(a A, tail Stream[A]) io.IO[Stream[B]] {
			zero2, stm2 := f(a, zero)
			return io.IO[Stream[B]](StateFlatMapWithFinish(tail, zero2, IOAndThen(prefix, stm2), f, onFinish))
		},
		/*onEmpty */ func(tail Stream[A]) io.IO[Stream[B]] {
			return io.IO[Stream[B]](StateFlatMapWithFinish(tail, zero, prefix, f, onFinish))
		},
		/*onError */ func(err error) io.IO[Stream[B]] {
			return io.Fail[Stream[B]](err)
		},
	)
	return IOStream[B](res)
}

// Filter leaves in the stream only the elements that satisfy the given predicate.
func Filter[A any](stm Stream[A], predicate func(A) bool) IOStream[A] {
	res := StreamMatch[A, Stream[A]](
		stm,
		/*onFinish*/ func() io.IO[Stream[A]] {
			return io.IO[Stream[A]](EmptyIO[A]())
		},
		/*onValue */ func(a A, tail Stream[A]) io.IO[Stream[A]] {
			if predicate(a) {
				return io.Lift(AndThen(Lift(a), tail))
			} else {
				return io.Lift(tail)
			}
		},
		/*onEmpty */ func(tail Stream[A]) io.IO[Stream[A]] {
			return io.Lift(tail)
		},
		/*onError */ func(err error) io.IO[Stream[A]] {
			return io.Fail[Stream[A]](err)
		},
	)
	return IOStream[A](res)
}

// Sum is a pipe that returns a stream of 1 element that is
// the sum of all elements of the original stream.
func Sum[A fun.Number](sa Stream[A]) IOStream[A] {
	h := stream.Head(stream.Sum(sa.init.sa))
	hf := io.Finally(h, sa.init.finalizer)
	hfs := stream.Eval(hf)
	return option.Match(sa.cont,
		func(sa2 IOStream[A]) IOStream[A] {
			iosa2 := io.IO[Stream[A]](sa2)
			return IOStream[A](io.FlatMap(hf,
				func(sum1 A) io.IO[Stream[A]] {
					return io.FlatMap(iosa2,
						func(s2 Stream[A]) io.IO[Stream[A]] {
							sum2 := io.IO[Stream[A]](Sum(s2))
							return io.Map(sum2, func(stm3 Stream[A]) Stream[A] {
								return Map(stm3, func(el A) A {
									return sum1 + el
								})
							})
						})
				}))
		},
		func() IOStream[A] { return LiftStreamIO(hfs) },
	)
}

// Sum2 is a pipe that returns a stream of 1 element that is sum of all elements of the original stream.
// It's another implementation of the same logic as in Sum.
func Sum2[A fun.Number](sa Stream[A]) IOStream[A] {
	var zero A
	emptyIOA := EmptyIO[A]()
	return StateFlatMapWithFinish(sa,
		zero,
		emptyIOA,
		func(a A, s A) (A, IOStream[A]) {
			return s + a, emptyIOA
		},
		LiftIO[A],
	)
}

// Len is a pipe that returns a stream of 1 element that is the count of elements of the original stream.
func Len[A any](sa Stream[A]) IOStream[int] {
	return Sum(Map(sa, fun.Const[A](1)))
}

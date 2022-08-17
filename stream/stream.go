// Package stream provides a way to construct data processing streams
// from smaller pieces.
// The design is inspired by fs2 Scala library.
package stream

import (
	"github.com/primetalk/goio/either"
	"github.com/primetalk/goio/fun"
	"github.com/primetalk/goio/io"
	"github.com/primetalk/goio/slice"
)

// Stream is modelled as a function that performs a single step in the state machine.
type Stream[A any] io.IO[StepResult[A]]

// StepResult[A] represents the result of a single step in the step machine.
// It might be one of - empty, new value, or finished.
// The step result also returns the continuation of the stream.
type StepResult[A any] struct {
	Value        A
	HasValue     bool // models "Option[A]"
	Continuation Stream[A]
	IsFinished   bool // true when stream has completed
}

// NewStepResult constructs StepResult that has one value.
func NewStepResult[A any](value A, continuation Stream[A]) StepResult[A] {
	return StepResult[A]{
		Value:        value,
		HasValue:     true,
		Continuation: continuation,
		IsFinished:   false,
	}
}

// NewStepResultEmpty constructs an empty StepResult.
func NewStepResultEmpty[A any](continuation Stream[A]) StepResult[A] {
	return StepResult[A]{
		HasValue:     false,
		Continuation: continuation,
		IsFinished:   false,
	}
}

// NewStepResultFinished constructs a finished StepResult.
// The continuation will be empty as well.
func NewStepResultFinished[A any]() StepResult[A] {
	return StepResult[A]{
		IsFinished:   true,
		Continuation: Empty[A](),
	}
}

// MapEval maps the values of the stream. The provided function returns an IO.
func MapEval[A any, B any](stm Stream[A], f func(a A) io.IO[B]) Stream[B] {
	return Stream[B](io.FlatMap(
		io.IO[StepResult[A]](stm),
		func(sra StepResult[A]) io.IO[StepResult[B]] {
			if sra.IsFinished {
				return io.Lift(NewStepResultFinished[B]())
			} else if sra.HasValue {
				iob := f(sra.Value)
				return io.Map(iob, func(b B) StepResult[B] {
					return NewStepResult(b, MapEval(sra.Continuation, f))
				})
			} else {
				return io.Lift(
					NewStepResultEmpty(MapEval(sra.Continuation, f)),
				)
			}
		}))
}

// Map converts values of the stream.
func Map[A any, B any](stm Stream[A], f func(a A) B) Stream[B] {
	return MapEval(stm, func(a A) io.IO[B] { return io.Lift(f(a)) })
}

// MapPipe creates a pipe that maps one stream through the provided function.
func MapPipe[A any, B any](f func(a A) B) Pipe[A, B] {
	return func(sa Stream[A]) Stream[B] {
		return Map(sa, f)
	}
}

// AndThen appends another stream after the end of the first one.
func AndThen[A any](stm1 Stream[A], stm2 Stream[A]) Stream[A] {
	return AndThenLazy(stm1, func() Stream[A] { return stm2 })
}

// AndThenLazy appends another stream. The other stream is constructed lazily.
func AndThenLazy[A any](stm1 Stream[A], stm2 func() Stream[A]) Stream[A] {
	return Stream[A](io.FlatMap(
		io.IO[StepResult[A]](stm1),
		func(sra StepResult[A]) io.IO[StepResult[A]] {
			if sra.IsFinished {
				return io.IO[StepResult[A]](stm2())
			} else {
				return io.Lift(StepResult[A]{
					Value:        sra.Value,
					Continuation: AndThenLazy(sra.Continuation, stm2),
					HasValue:     sra.HasValue,
				})
			}
		}))

}

// FlatMap constructs a new stream by concatenating all substreams, produced by f
// from elements of the original stream.
func FlatMap[A any, B any](stm Stream[A], f func(a A) Stream[B]) Stream[B] {
	return Stream[B](io.FlatMap(
		io.IO[StepResult[A]](stm),
		func(sra StepResult[A]) io.IO[StepResult[B]] {
			if sra.IsFinished {
				return io.Lift(NewStepResultFinished[B]())
			} else if sra.HasValue {
				stmb1 := f(sra.Value)
				stmb := AndThenLazy(stmb1, func() Stream[B] { return FlatMap(sra.Continuation, f) })
				return io.IO[StepResult[B]](stmb)
			} else {
				return io.Lift(NewStepResultEmpty(FlatMap(sra.Continuation, f)))
			}
		}))
}

// FoldToGoResult converts a stream into a stream of go results.
// All go results will be non-error except probably the last one.
func FoldToGoResult[A any](stm Stream[A]) Stream[io.GoResult[A]] {
	return Stream[io.GoResult[A]](
		io.Map(io.FoldToGoResult(io.IO[StepResult[A]](stm)), func(gra io.GoResult[StepResult[A]]) StepResult[io.GoResult[A]] {
			if gra.Error == nil {
				sra := gra.Value
				if sra.IsFinished {
					return NewStepResultFinished[io.GoResult[A]]()
				} else if sra.HasValue {
					return NewStepResult(io.NewGoResult(sra.Value), FoldToGoResult(gra.Value.Continuation))
				} else {
					return NewStepResultEmpty(FoldToGoResult(sra.Continuation))
				}
			} else {
				return NewStepResult(io.NewFailedGoResult[A](gra.Error), Empty[io.GoResult[A]]())
			}
		}),
	)
}

// UnfoldGoResult converts a stream of GoResults back to normal stream.
// On the first encounter of Error, the stream fails.
// default value for `onFailure`  - `Fail[A]`.
func UnfoldGoResult[A any](stm Stream[io.GoResult[A]], onFailure func(err error) Stream[A]) Stream[A] {
	return Stream[A](
		FlatMap(stm,
			func(gra io.GoResult[A]) Stream[A] {
				if gra.Error == nil {
					return Lift(gra.Value)
				} else {
					return onFailure(gra.Error)
				}
			},
		),
	)
}

// FlatMapPipe creates a pipe that flatmaps one stream through the provided function.
func FlatMapPipe[A any, B any](f func(a A) Stream[B]) Pipe[A, B] {
	return func(sa Stream[A]) Stream[B] {
		return FlatMap(sa, f)
	}
}

// Flatten simplifies a stream of streams to just the stream of values by concatenating all
// inner streams.
func Flatten[A any](stm Stream[Stream[A]]) Stream[A] {
	return FlatMap(stm, fun.Identity[Stream[A]])
}

// StateFlatMap maintains state along the way.
func StateFlatMap[A any, B any, S any](stm Stream[A], zero S, f func(a A, s S) io.IO[fun.Pair[S, Stream[B]]]) Stream[B] {
	return StateFlatMapWithFinish(stm, zero, f, func(S) Stream[B] { return Empty[B]() })
}

// StateFlatMapWithFinish maintains state along the way.
// When the source stream finishes, it invokes onFinish with the last state.
func StateFlatMapWithFinish[A any, B any, S any](stm Stream[A],
	zero S,
	f func(a A, s S) io.IO[fun.Pair[S, Stream[B]]],
	onFinish func(s S) Stream[B]) Stream[B] {
	res := io.FlatMap(
		io.IO[StepResult[A]](stm),
		func(sra StepResult[A]) (iores io.IO[StepResult[B]]) {
			if sra.IsFinished {
				iores = io.Lift(NewStepResultEmpty(onFinish(zero)))
			} else if sra.HasValue {
				iop := f(sra.Value, zero)
				iores = io.FlatMap(iop, func(p fun.Pair[S, Stream[B]]) io.IO[StepResult[B]] {
					st, stmb1 := p.V1, p.V2
					stmb := AndThenLazy(stmb1, func() Stream[B] { return StateFlatMapWithFinish(sra.Continuation, st, f, onFinish) })
					return io.IO[StepResult[B]](stmb)
				})
			} else {
				iores = io.Lift(NewStepResultEmpty(StateFlatMapWithFinish(sra.Continuation, zero, f, onFinish)))
			}
			return
		})
	return Stream[B](res)
}

// Filter leaves in the stream only the elements that satisfy the given predicate.
func Filter[A any](stm Stream[A], predicate func(A) bool) Stream[A] {
	return Stream[A](io.Map(
		io.IO[StepResult[A]](stm),
		func(sra StepResult[A]) StepResult[A] {
			if sra.IsFinished {
				return sra
			} else {
				cont := Filter(sra.Continuation, predicate)
				if sra.HasValue && predicate(sra.Value) {
					return NewStepResult(sra.Value, cont)
				} else {
					return NewStepResultEmpty(cont)
				}
			}
		}))
}

// Filter leaves in the stream only the elements
// that do not satisfy the given predicate.
func FilterNot[A any](stm Stream[A], predicate func(A) bool) Stream[A] {
	return Filter(stm, func(a A) bool {
		return !predicate(a)
	})
}

// Sum is a pipe that returns a stream of 1 element that is sum of all elements of the original stream.
func Sum[A slice.Number](sa Stream[A]) Stream[A] {
	var zero A
	return StateFlatMapWithFinish(sa, zero,
		func(a A, s A) io.IO[fun.Pair[A, Stream[A]]] {
			return io.Lift(fun.NewPair(s+a, Empty[A]()))
		},
		func(lastState A) Stream[A] {
			return Lift(lastState)
		})
}

// Len is a pipe that returns a stream of 1 element that is the count of elements of the original stream.
func Len[A any](sa Stream[A]) Stream[int] {
	return Sum(Map(sa, fun.Const[A](1)))
}

// ChunkN groups elements by n and produces a stream of slices.
func ChunkN[A any](n int) func(sa Stream[A]) Stream[[]A] {
	return func(sa Stream[A]) Stream[[]A] {
		return StateFlatMapWithFinish(sa, make([]A, 0, n),
			func(a A, as []A) io.IO[fun.Pair[[]A, Stream[[]A]]] {
				if len(as) == n-1 {
					return io.Lift(fun.NewPair(make([]A, 0, n), Lift(append(as, a))))
				} else {
					return io.Lift(fun.NewPair(append(as, a), Empty[[]A]()))
				}
			},
			func(as []A) Stream[[]A] {
				return Lift(as)
			},
		)
	}
}

// Fail returns a stream that fails immediately.
func Fail[A any](err error) Stream[A] {
	return Eval(io.Fail[A](err))
}

// GroupBy collects group by a user-provided key. Whenever a new key is encountered,
// the previous group is emitted.
// When the original stream finishes, the last group is emitted.
func GroupBy[A any, K comparable](stm Stream[A], key func(A) K) Stream[fun.Pair[K, []A]] {
	zero := fun.Pair[K, []A]{
		V2: []A{},
	}
	return StateFlatMapWithFinish(stm,
		zero,
		func(a A, s fun.Pair[K, []A]) io.IO[fun.Pair[fun.Pair[K, []A], Stream[fun.Pair[K, []A]]]] {
			return io.Pure(func() (result fun.Pair[fun.Pair[K, []A], Stream[fun.Pair[K, []A]]]) {
				aKey := key(a)
				if len(s.V2) == 0 || s.V1 == aKey {
					result = fun.Pair[fun.Pair[K, []A], Stream[fun.Pair[K, []A]]]{
						V1: fun.Pair[K, []A]{
							V1: aKey,
							V2: append(s.V2, a),
						},
						V2: Empty[fun.Pair[K, []A]](),
					}
				} else {
					result = fun.Pair[fun.Pair[K, []A], Stream[fun.Pair[K, []A]]]{
						V1: fun.Pair[K, []A]{
							V1: aKey,
							V2: []A{a},
						},
						V2: Lift(s),
					}
				}
				return
			})
		},
		func(last fun.Pair[K, []A]) Stream[fun.Pair[K, []A]] {
			if len(last.V2) == 0 {
				return Empty[fun.Pair[K, []A]]()
			} else {
				return Lift(last)
			}
		},
	)
}

// GroupByEval collects group by a user-provided key (which is evaluated as IO).
// Whenever a new key is encountered, the previous group is emitted.
// When the original stream finishes, the last group is emitted.
func GroupByEval[A any, K comparable](stm Stream[A], keyIO func(A) io.IO[K]) Stream[fun.Pair[K, []A]] {
	zero := fun.Pair[K, []A]{
		V2: []A{},
	}
	return StateFlatMapWithFinish(stm,
		zero,
		func(a A, s fun.Pair[K, []A]) io.IO[fun.Pair[fun.Pair[K, []A], Stream[fun.Pair[K, []A]]]] {
			return io.Map(keyIO(a), func(aKey K) (result fun.Pair[fun.Pair[K, []A], Stream[fun.Pair[K, []A]]]) {
				if len(s.V2) == 0 || s.V1 == aKey {
					result = fun.Pair[fun.Pair[K, []A], Stream[fun.Pair[K, []A]]]{
						V1: fun.Pair[K, []A]{
							V1: aKey,
							V2: append(s.V2, a),
						},
						V2: Empty[fun.Pair[K, []A]](),
					}
				} else {
					result = fun.Pair[fun.Pair[K, []A], Stream[fun.Pair[K, []A]]]{
						V1: fun.Pair[K, []A]{
							V1: aKey,
							V2: []A{a},
						},
						V2: Lift(s),
					}
				}
				return
			})
		},
		func(last fun.Pair[K, []A]) Stream[fun.Pair[K, []A]] {
			if len(last.V2) == 0 {
				return Empty[fun.Pair[K, []A]]()
			} else {
				return Lift(last)
			}
		},
	)
}

// FanOut distributes the same element to all handlers.
// Stream failure is also distribured to all handlers.
func FanOut[A any, B any](stm Stream[A], handlers ...func(Stream[A]) io.IO[B]) io.IO[[]B] {
	gra := FoldToGoResult(stm)
	var channels []chan io.GoResult[A]
	// NB: sideeffectful mapping:
	ios := slice.Map(handlers, func(handler func(Stream[A]) io.IO[B]) io.IO[B] {
		ch := make(chan io.GoResult[A])
		channels = append(channels, ch)
		stmCh := UnfoldGoResult(FromChannel(ch), Fail[A])
		return handler(stmCh)
	})
	channelsIn := slice.Map(channels, func(ch chan io.GoResult[A]) chan<- io.GoResult[A] {
		return ch
	})
	toChannelsIO := ToChannels(gra, channelsIn...)
	toChannelsIOCompatible := io.Map(toChannelsIO, either.Left[fun.Unit, []B])
	iosParallelIO := io.Parallel(ios...)
	iosParallelIOCompatible := io.Map(iosParallelIO, either.Right[fun.Unit, []B])
	both := io.Parallel(toChannelsIOCompatible, iosParallelIOCompatible)
	onlyRight := io.Map(both, func(eithers []either.Either[fun.Unit, []B]) []B {
		return slice.Flatten(slice.Collect(eithers, either.GetRight[fun.Unit, []B]))
	})
	return onlyRight
}

// FoldLeftEval aggregates stream in a more simple way than StateFlatMap.
func FoldLeftEval[A any, B any](stm Stream[A], zero B, combine func(B, A) io.IO[B]) io.IO[B] {
	return Head(
		StateFlatMapWithFinish(stm, zero,
			func(a A, b B) io.IO[fun.Pair[B, Stream[B]]] {
				return io.Map(combine(b, a), func(b B) fun.Pair[B, Stream[B]] {
					return fun.NewPair(b, Empty[B]())
				})
			},
			func(b B) Stream[B] {
				return Lift(b)
			},
		),
	)
}

// FoldLeft aggregates stream in a more simple way than StateFlatMap.
func FoldLeft[A any, B any](stm Stream[A], zero B, combine func(B, A) B) io.IO[B] {
	return FoldLeftEval(stm, zero, func(b B, a A) io.IO[B] {
		return io.Pure(func() B {
			return combine(b, a)
		})
	})
}
